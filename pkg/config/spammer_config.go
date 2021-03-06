package config

import (
	flag "github.com/spf13/pflag"
)

const (
	// the target address of the spam
	CfgSpammerAddress = "spammer.address"
	// the message to embed within the spam transactions
	CfgSpammerMessage = "spammer.message"
	// the tag of the transaction
	CfgSpammerTag = "spammer.tag"
	// the tag of the transaction if the semi-lazy pool is used (uses "tag" if empty)
	CfgSpammerTagSemiLazy = "spammer.tagSemiLazy"
	// workers remains idle for a while when cpu usage gets over this limit (0 = disable)
	CfgSpammerCPUMaxUsage = "spammer.cpuMaxUsage"
	// the rate limit for the spammer (0 = no limit)
	CfgSpammerTPSRateLimit = "spammer.tpsRateLimit"
	// the size of the spam bundles
	CfgSpammerBundleSize = "spammer.bundleSize"
	// should be spammed with value bundles
	CfgSpammerValueSpam = "spammer.valueSpam"
	// the amount of parallel running spammers
	CfgSpammerWorkers = "spammer.workers"
	// the maximum amount of tips in the semi-lazy tip-pool before the spammer tries to reduce these (0 = disable)
	// this is used to support the network if someone attacks the tangle by spamming almost lazy tips
	CfgSpammerSemiLazyTipsLimit = "spammer.semiLazyTipsLimit"
)

func init() {
	flag.String(CfgSpammerAddress, "HORNET99INTEGRATED99SPAMMER999999999999999999999999999999999999999999999999999999", "the target address of the spam")
	flag.String(CfgSpammerMessage, "Spamming with HORNET tipselect", "the message to embed within the spam transactions")
	flag.String(CfgSpammerTag, "HORNET99SPAMMER999999999999", "the tag of the transaction")
	flag.String(CfgSpammerTagSemiLazy, "", "the tag of the transaction if the semi-lazy pool is used (uses \"tag\" if empty)")
	flag.Float64(CfgSpammerCPUMaxUsage, 0.50, "workers remains idle for a while when cpu usage gets over this limit (0 = disable)")
	flag.Float64(CfgSpammerTPSRateLimit, 0.10, "the rate limit for the spammer (0 = no limit)")
	flag.Int(CfgSpammerBundleSize, 1, "the size of the spam bundles")
	flag.Bool(CfgSpammerValueSpam, false, "should be spammed with value bundles")
	flag.Int(CfgSpammerWorkers, 1, "the amount of parallel running spammers")
	flag.Int(CfgSpammerSemiLazyTipsLimit, 20, "the maximum amount of tips in the semi-lazy tip-pool before the spammer tries to reduce these (0 = disable)")
}
