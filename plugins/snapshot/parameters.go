package snapshot

import flag "github.com/spf13/pflag"

func init() {
	flag.Bool("pruning.enabled", true, "Delete old transactions from database")
	flag.Int("pruning.delay", 40000, "Amount of milestone transactions to keep in the database")

	flag.Bool("localSnapshots.enabled", true, "Enable local snapshots")
	flag.Int("localSnapshots.depth", 100, "Amount of seen milestones to record in the snapshot file")
	flag.Int("localSnapshots.intervalSynced", 10, "Interval, in milestone transactions, at which snapshot files are created if the ledger is fully synchronized")
	flag.Int("localSnapshots.intervalUnsynced", 1000, "Interval, in milestone transactions, at which snapshot files are created if the ledger is not fully synchronized")
	flag.String("localSnapshots.path", "latest-export.gz.bin", "Path to the local snapshot file")

	flag.Bool("globalSnapshot.load", false, "Load global snapshot")
	flag.String("globalSnapshot.path", "snapshotMainnet.txt", "Path to the global snapshot file")
	flag.StringSlice("globalSnapshot.spentAddressesPaths", []string{"previousEpochsSpentAddresses1.txt", "previousEpochsSpentAddresses2.txt", "previousEpochsSpentAddresses3.txt"}, "Paths to the spent addresses files")
	flag.Int("globalSnapshot.index", 1050000, "Milestone index of the global snapshot")

	flag.String("privateTangle.ledgerStatePath", "balances.txt", "Path to the ledger state file for your private tangle")
}
