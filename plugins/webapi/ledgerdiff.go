package webapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gohornet/hornet/packages/model/milestone_index"
	"github.com/gohornet/hornet/packages/model/tangle"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

func init() {
	addEndpoint("getLedgerDiff", getLedgerDiff, implementedAPIcalls)
	addEndpoint("getLedgerDiffExt", getLedgerDiffExt, implementedAPIcalls)
}

func getLedgerDiff(i interface{}, c *gin.Context) {
	ld := &GetLedgerDiff{}
	e := ErrorReturn{}

	err := mapstructure.Decode(i, ld)
	if err != nil {
		e.Error = "Internal error"
		c.JSON(http.StatusInternalServerError, e)
		return
	}

	smi := tangle.GetSolidMilestoneIndex()
	requestedIndex := milestone_index.MilestoneIndex(ld.MilestoneIndex)
	if requestedIndex > smi {
		e.Error = fmt.Sprintf("Invalid milestone index supplied, lsmi is %d", smi)
		c.JSON(http.StatusBadRequest, e)
		return
	}

	ldr := &GetLedgerDiffReturn{}

	diff, err := tangle.GetLedgerDiffForMilestone(requestedIndex)
	if err != nil {
		e.Error = "Internal error"
		c.JSON(http.StatusInternalServerError, e)
		return
	}

	ldr.Diff = diff
	ldr.MilestoneIndex = ld.MilestoneIndex

	c.JSON(http.StatusOK, ldr)
}

func getLedgerDiffExt(i interface{}, c *gin.Context) {
	ld := &GetLedgerDiff{}
	e := ErrorReturn{}

	err := mapstructure.Decode(i, ld)
	if err != nil {
		e.Error = "Internal error"
		c.JSON(http.StatusInternalServerError, e)
		return
	}

	smi := tangle.GetSolidMilestoneIndex()
	requestedIndex := milestone_index.MilestoneIndex(ld.MilestoneIndex)
	if requestedIndex > smi {
		e.Error = fmt.Sprintf("Invalid milestone index supplied, lsmi is %d", smi)
		c.JSON(http.StatusBadRequest, e)
		return
	}

	ldr := &GetLedgerDiffExtReturn{}

	diff, err := tangle.GetLedgerDiffForMilestoneExt(requestedIndex)
	if err != nil {
		e.Error = "Internal error"
		c.JSON(http.StatusInternalServerError, e)
		return
	}

	for address, bundles := range diff {
		for _, bundle := range bundles {
			var valueChange int64 = 0
			for _, tx := range bundle.GetTransactions() {
				if tx.Tx.Address == address {
					valueChange += tx.Tx.Value
				}
			}
			ldr.Diff[address] = []StringDiff{}
			ldr.Diff[address] = append(ldr.Diff[address], StringDiff{bundle.GetTail().GetHash(), valueChange})
		}
	}
	ldr.MilestoneIndex = ld.MilestoneIndex

	c.JSON(http.StatusOK, ldr)
}
