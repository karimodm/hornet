package tangle

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/iotaledger/hive.go/parameter"
	"github.com/labstack/gommon/log"

	"github.com/gohornet/hornet/packages/compressed"

	"github.com/pkg/errors"

	"github.com/gohornet/hornet/packages/database"
	"github.com/gohornet/hornet/packages/model/milestone_index"
	"github.com/gohornet/hornet/packages/typeutils"
	"github.com/iotaledger/iota.go/trinary"
)

var (
	ledgerDatabase                database.Database
	ledgerDatabaseTransactionLock sync.RWMutex
	ledgerMilestoneIndex          milestone_index.MilestoneIndex
	balancePrefix                 = []byte("balance")
	diffPrefix                    = []byte("diff")
)

func ReadLockLedger() {
	ledgerDatabaseTransactionLock.RLock()
}

func ReadUnlockLedger() {
	ledgerDatabaseTransactionLock.RUnlock()
}

func WriteLockLedger() {
	ledgerDatabaseTransactionLock.Lock()
}

func WriteUnlockLedger() {
	ledgerDatabaseTransactionLock.Unlock()
}

func configureLedgerDatabase() {
	if db, err := database.Get(DBPrefixLedgerState); err != nil {
		panic(err)
	} else {
		ledgerDatabase = db
	}

	loadLSMIAsLSM := parameter.NodeConfig.GetBool("compass.loadLSMIAsLMI")
	err := readLedgerMilestoneIndexFromDatabase(loadLSMIAsLSM)
	if err != nil {
		panic(err)
	}
}

func databaseKeyForAddressBalance(address trinary.Hash) []byte {
	return append(balancePrefix, trinary.MustTrytesToBytes(address)...)
}

func databaseKeyPrefixForLedgerDiff(milestoneIndex milestone_index.MilestoneIndex) []byte {
	return append(diffPrefix, databaseKeyForMilestoneIndex(milestoneIndex)...)
}

func databaseKeyForLedgerDiffAndAddress(milestoneIndex milestone_index.MilestoneIndex, address trinary.Hash) []byte {
	return append(databaseKeyPrefixForLedgerDiff(milestoneIndex), trinary.MustTrytesToBytes(address)...)
}

func bytesFromBalance(balance uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, balance)
	return bytes
}

func balanceFromBytes(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

func bytesFromDiff(diff int64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(diff))
	return bytes
}

func diffFromBytes(bytes []byte) int64 {
	return int64(balanceFromBytes(bytes))
}

func entryForMilestoneIndex(index milestone_index.MilestoneIndex) database.Entry {
	return database.Entry{
		Key:   typeutils.StringToBytes("ledgerMilestoneIndex"),
		Value: bytesFromMilestoneIndex(index),
	}
}

func readLedgerMilestoneIndexFromDatabase(setLSMIAsLMI bool) error {

	ReadLockLedger()
	defer ReadUnlockLedger()

	entry, err := ledgerDatabase.Get(typeutils.StringToBytes("ledgerMilestoneIndex"))
	if err != nil {
		if err == database.ErrKeyNotFound {
			return nil
		} else {
			return errors.Wrap(NewDatabaseError(err), "failed to retrieve ledger milestone index")
		}
	}

	ledgerMilestoneIndex = milestoneIndexFromBytes(entry.Value)

	// Set the solid milestone index based on the ledger milestone
	setSolidMilestoneIndex(ledgerMilestoneIndex)
	if setLSMIAsLMI && ledgerMilestoneIndex != 0 {
		solidMsBundle, err := GetMilestone(ledgerMilestoneIndex)
		if err != nil {
			return errors.Wrap(NewDatabaseError(err), "failed to retrieve ledger milestone bundle")
		}
		if solidMsBundle != nil {
			SetLatestMilestone(solidMsBundle)
		}
	}

	return nil
}

func GetBalanceForAddressWithoutLocking(address trinary.Hash) (uint64, milestone_index.MilestoneIndex, error) {

	entry, err := ledgerDatabase.Get(databaseKeyForAddressBalance(address))
	if err != nil {
		if err == database.ErrKeyNotFound {
			return 0, ledgerMilestoneIndex, nil
		} else {
			return 0, ledgerMilestoneIndex, errors.Wrap(NewDatabaseError(err), "failed to retrieve balance")
		}
	}

	return balanceFromBytes(entry.Value), ledgerMilestoneIndex, err
}

func GetBalanceForAddress(address trinary.Hash) (uint64, milestone_index.MilestoneIndex, error) {

	ReadLockLedger()
	defer ReadUnlockLedger()

	return GetBalanceForAddressWithoutLocking(address)
}

func DeleteLedgerDiffForMilestone(index milestone_index.MilestoneIndex) error {
	WriteLockLedger()
	defer WriteUnlockLedger()
	return ledgerDatabase.DeletePrefix(databaseKeyPrefixForLedgerDiff(index))
}

func GetLedgerDiffForMilestone(index milestone_index.MilestoneIndex) (map[trinary.Hash]int64, error) {

	ReadLockLedger()
	defer ReadUnlockLedger()

	diff := make(map[trinary.Hash]int64)

	err := ledgerDatabase.ForEachPrefix(databaseKeyPrefixForLedgerDiff(index), func(entry database.Entry) (stop bool) {
		address := trinary.MustBytesToTrytes(entry.Key, 81)
		diff[address] = diffFromBytes(entry.Value)
		return false
	})

	if err != nil {
		return nil, err
	}

	var diffSum int64
	for _, change := range diff {
		diffSum += change
	}

	if diffSum != 0 {
		panic(fmt.Sprintf("GetLedgerDiffForMilestone(): Ledger diff for milestone %d does not sum up to zero", index))
	}

	return diff, nil
}

func GetLedgerDiffForMilestoneExt(index milestone_index.MilestoneIndex) (map[trinary.Hash][]*Bundle, error) {

	ReadLockLedger()
	defer ReadUnlockLedger()

	diff := make(map[trinary.Hash][]*Bundle)

	err := ledgerDatabase.ForEachPrefix(databaseKeyPrefixForLedgerDiff(index), func(entry database.Entry) (stop bool) {
		address := trinary.MustBytesToTrytes(entry.Key, 81)
		diff[address] = getConeForAddress(index, address)
		return false
	})

	return diff, err
}

func getConeForAddress(index milestone_index.MilestoneIndex, findAddress string) []*Bundle {

	milestoneBundle, _ := GetMilestone(index)
	milestoneTail := milestoneBundle.GetTail()
	txsToConfirm := make(map[string]struct{})
	txsToTraverse := make(map[string]struct{})
	totalLedgerChanges := make(map[string]int64)
	txsToTraverse[milestoneTail.GetHash()] = struct{}{}

	var filteredBundles []*Bundle

	// Collect all tx to check by traversing the tangle
	// Loop as long as new transactions are added in every loop cycle
	for len(txsToTraverse) != 0 {

		for txHash := range txsToTraverse {
			delete(txsToTraverse, txHash)

			if _, checked := txsToConfirm[txHash]; checked {
				// Tx was already checked => ignore
				continue
			}

			if SolidEntryPointsContain(txHash) {
				// Ignore solid entry points (snapshot milestone included)
				continue
			}

			tx, _ := GetTransaction(txHash)
			if tx == nil {
				log.Panicf("confirmMilestone: Transaction not found: %v", txHash)
			}

			confirmed, at := tx.GetConfirmed()
			if confirmed {
				if at > index {
					log.Panicf("transaction %s was already confirmed by a newer milestone %d", tx.GetHash(), at)
				}

				// Tx is already confirmed by another milestone => ignore
				if at < index {
					continue
				}

				// If confirmationIndex == milestoneIndex,
				// we have to walk the ledger changes again (for re-applying the ledger changes after shutdown)
			}

			// Mark the approvees to be traversed
			txsToTraverse[tx.GetTrunk()] = struct{}{}
			txsToTraverse[tx.GetBranch()] = struct{}{}

			if !tx.IsTail() {
				continue
			}

			bundleBucket, err := GetBundleBucket(tx.Tx.Bundle)
			if err != nil {
				log.Panicf("confirmMilestone: BundleBucket not found: %v, Error: %v", tx.Tx.Bundle, err)
			}

			bundle := bundleBucket.GetBundleOfTailTransaction(txHash)
			if bundle == nil {
				log.Panicf("confirmMilestone: Tx: %v, Bundle not found: %v", txHash, tx.Tx.Bundle)
			}

			if !bundle.IsValid() {
				log.Panicf("confirmMilestone: Tx: %v, Bundle not valid: %v", txHash, tx.Tx.Bundle)
			}

			if !bundle.IsComplete() {
				log.Panicf("confirmMilestone: Tx: %v, Bundle not complete: %v", txHash, tx.Tx.Bundle)
			}

			ledgerChanges, isValueSpamBundle := bundle.GetLedgerChanges()
			if !isValueSpamBundle {
				for address, change := range ledgerChanges {
					if address == findAddress {
						filteredBundles = append(filteredBundles, bundle)
					}
					totalLedgerChanges[address] += change
				}
			}

			// we only add the tail transaction to the txsToConfirm set, in order to not
			// accidentally skip cones, in case the other transactions (non-tail) of the bundle do not
			// reference the same trunk transaction (as seen from the PoV of the bundle).
			// if we wouldn't do it like this, we have a high chance of computing an
			// inconsistent ledger state.
			txsToConfirm[txHash] = struct{}{}
		}
	}

	return filteredBundles
}

func ApplyLedgerDiff(diff map[trinary.Hash]int64, index milestone_index.MilestoneIndex) error {

	var diffEntries []database.Entry
	var balanceChanges []database.Entry
	var emptyAddresses []database.Key

	var diffSum int64

	for address, change := range diff {

		balance, _, err := GetBalanceForAddressWithoutLocking(address)
		if err != nil {
			panic(fmt.Sprintf("GetBalanceForAddressWithoutLocking() returned error for address %s: %v", address, err))
		}

		newBalance := int64(balance) + change

		if newBalance < 0 {
			panic(fmt.Sprintf("Ledger diff for milestone %d creates negative balance for address %s: current %d, diff %d", index, address, balance, change))
		} else if newBalance > 0 {
			balanceChanges = append(balanceChanges, database.Entry{
				Key:   databaseKeyForAddressBalance(address),
				Value: bytesFromBalance(uint64(newBalance)),
			})
		} else {
			// Balance is zero, so we can remove this address from the ledger
			emptyAddresses = append(emptyAddresses, databaseKeyForAddressBalance(address))
		}

		diffEntries = append(diffEntries, database.Entry{
			Key:   databaseKeyForLedgerDiffAndAddress(index, address),
			Value: bytesFromDiff(change),
		})

		diffSum += change
	}

	if diffSum != 0 {
		panic(fmt.Sprintf("Ledger diff for milestone %d does not sum up to zero", index))
	}

	entries := balanceChanges
	entries = append(entries, diffEntries...)
	entries = append(entries, entryForMilestoneIndex(index))
	deletions := emptyAddresses

	// Now batch insert/delete all entries
	if err := ledgerDatabase.Apply(entries, deletions); err != nil {
		return errors.Wrap(NewDatabaseError(err), "failed to store ledger diff")
	}

	ledgerMilestoneIndex = index
	return nil
}

func StoreBalancesInDatabase(balances map[trinary.Hash]uint64, index milestone_index.MilestoneIndex) error {

	WriteLockLedger()
	defer WriteUnlockLedger()

	var entries []database.Entry
	var deletions []database.Key

	for address, balance := range balances {
		key := databaseKeyForAddressBalance(address)
		if balance == 0 {
			deletions = append(deletions, key)
		} else {
			entries = append(entries, database.Entry{
				Key:   key,
				Value: bytesFromBalance(balance),
			})
		}
	}

	entries = append(entries, entryForMilestoneIndex(index))

	// Now batch insert/delete all entries
	if err := ledgerDatabase.Apply(entries, deletions); err != nil {
		return errors.Wrap(NewDatabaseError(err), "failed to store ledger state")
	}

	ledgerMilestoneIndex = index
	return nil
}

func GetAllBalances() (map[trinary.Hash]uint64, milestone_index.MilestoneIndex, error) {

	ReadLockLedger()
	defer ReadUnlockLedger()

	balances := make(map[trinary.Hash]uint64)

	err := ledgerDatabase.ForEachPrefix(balancePrefix, func(entry database.Entry) (stop bool) {
		address := trinary.MustBytesToTrytes(entry.Key, 81)
		balances[address] = balanceFromBytes(entry.Value)
		return false
	})

	if err != nil {
		return nil, ledgerMilestoneIndex, err
	}

	var total uint64
	for _, value := range balances {
		total += value
	}

	if total != compressed.TOTAL_SUPPLY {
		panic(fmt.Sprintf("GetAllBalances() Total does not match supply: %d != %d", total, compressed.TOTAL_SUPPLY))
	}

	return balances, ledgerMilestoneIndex, err
}
