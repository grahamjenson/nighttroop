package alarms

import "testing"

// Each test is timelines run by the rig in rig_test.go, which explains how they are written. Readings and thresholds
// are in mg/dL, as they are stored: 56 is 3.1 mmol/L, 72 is 4.0, 180 is 10 and 270 is 15.

func TestNight(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "a night with every alarm on, at their example settings",
			config: "urgent_low 56, low 72 5m, high 180 30m, urgent_high 270, no_data 20m",
			timeline: `
				0m    100                                 // 5.5 mmol/L, in range
				5m    95
				10m   70                                  // under 4.0
				15m   65    -> low started                // and it has lasted 5 minutes
				20m   54    -> urgent_low started         active: urgent_low, low   alerting: urgent_low   // under 3.1, which quiets the low
				25m   50    -> urgent_low repeated
				30m   62    -> urgent_low cleared, low repeated   // treated, over 3.1 but still under 4.0
				35m   80    -> low cleared                active:
				40m   190                                 // a big meal, over 10
				55m   220
				70m   240   -> high started               // which has lasted 30 minutes
				75m   280   -> urgent_high started        // over 15
				80m   260   -> urgent_high cleared        active: high   alerting: high   // the high notified 10 minutes ago, so waits
				85m   150   -> high cleared
				90m   140                                 // the sensor stops sending
				109m  check
				110m  check -> no_data started            // 20 minutes without a reading
				115m  130   -> no_data cleared            active:
			`,
		},
	})
}

func TestUrgentLow(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:     "a reading under it starts it at once",
			config:   "urgent_low 56",
			timeline: `0m 55 -> urgent_low started   active: urgent_low`,
		},
		{
			name:     "a reading at it is not under it",
			config:   "urgent_low 56",
			timeline: `0m 56   active:`,
		},
		{
			name:     "3.1 mmol/L is 55.9 mg/dL, under the 56 mg/dL 3.1 is stored as",
			config:   "urgent_low 56",
			timeline: `0m 55.9 -> urgent_low started`,
		},
		{
			name:   "a reading over it clears it",
			config: "urgent_low 56",
			timeline: `
				0m 50 -> urgent_low started
				5m 57 -> urgent_low cleared   active:
			`,
		},
		{
			name:   "it repeats every 5 minutes while active, on readings or checks",
			config: "urgent_low 56",
			timeline: `
				0m     50     -> urgent_low started
				4m59s  check
				5m     check  -> urgent_low repeated
				7m     50
				10m    45     -> urgent_low repeated
			`,
		},
		{
			name:   "only a reading clears it, so without one it keeps repeating",
			config: "urgent_low 56",
			timeline: `
				0m   50     -> urgent_low started
				5m   check  -> urgent_low repeated
				10m  check  -> urgent_low repeated
				60m  check  -> urgent_low repeated
			`,
		},
		{
			name:     "turned off, it never starts",
			config:   "",
			timeline: `0m 40   active:`,
		},
		{
			name:   "a reading backfilled over it, older than the newest, does not clear it",
			config: "urgent_low 56",
			timeline: `
				0m 50            -> urgent_low started
				1m 100 at -1m    active: urgent_low
			`,
		},
		{
			name:   "a reading from a clock that is ahead is still the newest",
			config: "urgent_low 56",
			timeline: `
				0m 50          -> urgent_low started
				1m 100 at 2m   -> urgent_low cleared
			`,
		},
	})
}

func TestLow(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:     "one reading under it is not enough",
			config:   "low 72 5m",
			timeline: `0m 70   active:`,
		},
		{
			name:   "it starts on the reading that makes it last 5 minutes",
			config: "low 72 5m",
			timeline: `
				0m     70
				4m59s  70
				5m     70   -> low started   active: low
			`,
		},
		{
			name:   "time alone does not make it last, a reading must",
			config: "low 72 5m",
			timeline: `
				0m  70
				4m  70
				8m  check
				9m  70     -> low started
			`,
		},
		{
			name:   "a reading at or over it starts the wait again",
			config: "low 72 5m",
			timeline: `
				0m  70
				3m  72
				4m  70
				8m  70
				9m  70   -> low started
			`,
		},
		{
			name:   "readings far apart still make it last",
			config: "low 72 5m",
			timeline: `
				0m   70
				30m  70   -> low started
			`,
		},
		{
			name:     "with 0 minutes it starts at once",
			config:   "low 72 0m",
			timeline: `0m 71.9 -> low started`,
		},
		{
			name:   "it repeats every 15 minutes, and a reading at it clears it",
			config: "low 72 5m",
			timeline: `
				0m      70
				5m      70     -> low started
				19m59s  check
				20m     check  -> low repeated
				25m     72     -> low cleared   active:
			`,
		},
		{
			name:   "a late reading for an earlier time can make it last",
			config: "low 72 5m",
			timeline: `
				0m   70
				10m  70 at 5m   -> low started
			`,
		},
		{
			name:   "a reading backfilled over it, between the lows, means it has not lasted",
			config: "low 72 5m",
			timeline: `
				0m  70
				5m  70          -> low started
				6m  80 at 2m    -> low cleared
			`,
		},
	})
}

func TestPredictedLow(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "falling a mg/dL a minute, 20 minutes on is under 72 once it is, and it clears as the fall slows",
			config: "predicted_low 72 20m",
			timeline: `
				0m   100
				5m   95                                  // 75 in 20 minutes, not under 72
				10m  90    -> predicted_low started      // 70 in 20 minutes
				15m  90    -> predicted_low cleared      active:   // 10 over 15 minutes, so 77 in 20
			`,
		},
		{
			name:   "the trend is from the newest readings within 15 minutes, not an older fall",
			config: "predicted_low 72 20m",
			timeline: `
				0m   200
				20m  100
				25m  99    active:   // 1 over 5 minutes, so 79 in 20
			`,
		},
		{
			name:   "readings less than 5 minutes apart are too close to make a trend",
			config: "predicted_low 72 20m",
			timeline: `
				0m   100
				4m   80    active:
			`,
		},
		{
			name:   "a steady glucose under it is not falling, so is not predicted, but low",
			config: "predicted_low 72 20m",
			timeline: `
				0m   70
				10m  70    active:
			`,
		},
		{
			name:     "rising is not predicted",
			config:   "predicted_low 72 20m",
			timeline: "0m 60\n10m 70   active:",
		},
		{
			name:   "it warns before the low, which quiets it once it is active",
			config: "low 72 0m, predicted_low 72 20m",
			timeline: `
				0m   90
				10m  80    -> predicted_low started                   // 60 in 20 minutes
				15m  70    -> low started   active: low, predicted_low   alerting: low
			`,
		},
	})
}

func TestHigh(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "it starts once over it for 30 minutes",
			config: "high 180 30m",
			timeline: `
				0m   181
				15m  200
				29m  190
				30m  185   -> high started
			`,
		},
		{
			name:   "a reading at it is not over it",
			config: "high 180 30m",
			timeline: `
				0m   181
				30m  180   active:
			`,
		},
		{
			name:   "a dip under it starts the wait again",
			config: "high 180 30m",
			timeline: `
				0m   200
				20m  170
				25m  200
				54m  200
				55m  200   -> high started
			`,
		},
		{
			name:   "it repeats every 15 minutes, and a reading under it clears it",
			config: "high 180 30m",
			timeline: `
				0m   200
				30m  200   -> high started
				45m  200   -> high repeated
				50m  150   -> high cleared
			`,
		},
	})
}

func TestUrgentHigh(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "a reading over it starts it at once, it repeats every 5 minutes, and one at it clears it",
			config: "urgent_high 270",
			timeline: `
				0m  271    -> urgent_high started
				5m  check  -> urgent_high repeated
				6m  270    -> urgent_high cleared
			`,
		},
		{
			name:     "a reading at it is not over it",
			config:   "urgent_high 270",
			timeline: `0m 270   active:`,
		},
	})
}

func TestFastRise(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "rising 18 within 15 minutes starts it, and rising slower clears it",
			config: "fast_rise 18 15m",
			timeline: `
				0m   100
				5m   106
				10m  112
				15m  118    -> fast_rise started   active: fast_rise   // 18 since 15 minutes ago
				20m  120    -> fast_rise cleared   active:             // only 14 since 15 minutes ago
			`,
		},
		{
			name:   "17 within 15 minutes is not fast enough",
			config: "fast_rise 18 15m",
			timeline: `
				0m   100
				15m  117   active:
			`,
		},
		{
			name:   "it is measured from the oldest reading within the minutes, not an older one",
			config: "fast_rise 18 15m",
			timeline: `
				0m   100
				5m   140    -> fast_rise started   // a quick jump within 15 minutes
				20m  150    -> fast_rise cleared   // 10 since 5m, the oldest within 15 minutes
			`,
		},
		{
			name:   "one reading, or one after a gap longer than the minutes, has nothing to rise from",
			config: "fast_rise 18 15m",
			timeline: `
				0m   100   active:
				30m  200   active:
			`,
		},
		{
			name:   "it stays on without a new reading, as it is measured by readings",
			config: "fast_rise 18 15m",
			timeline: `
				0m   100
				10m  130    -> fast_rise started
				24m  check  active: fast_rise
				25m  check  -> fast_rise repeated
			`,
		},
		{
			name:     "falling does not start it",
			config:   "fast_rise 18 15m",
			timeline: "0m 150\n10m 100   active:",
		},
	})
}

func TestFastDrop(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "dropping 18 within 15 minutes starts it, and dropping slower clears it",
			config: "fast_drop 18 15m",
			timeline: `
				0m   150
				5m   144
				10m  138
				15m  132    -> fast_drop started   active: fast_drop
				20m  130    -> fast_drop cleared   active:
			`,
		},
		{
			name:     "rising does not start it",
			config:   "fast_drop 18 15m",
			timeline: "0m 100\n10m 150   active:",
		},
		{
			name:   "it alerts beside the low it is heading into, which does not quiet it",
			config: "low 72 0m, fast_drop 18 15m",
			timeline: `
				0m   100
				10m  70    -> low started, fast_drop started   active: low, fast_drop   alerting: low, fast_drop
			`,
		},
	})
}

func TestNoData(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "without any reading ever there is nothing missing",
			config: "no_data 20m",
			timeline: `
				0m   check
				60m  check   active:
			`,
		},
		{
			name:   "it starts 20 minutes after the newest reading, repeats every 15, and a reading clears it",
			config: "no_data 20m",
			timeline: `
				0m      100
				19m59s  check
				20m     check  -> no_data started
				34m     check
				35m     check  -> no_data repeated
				40m     100    -> no_data cleared   active:
			`,
		},
		{
			name:   "each reading starts the wait again",
			config: "no_data 20m",
			timeline: `
				0m   100
				15m  100
				34m  check
				35m  check  -> no_data started
			`,
		},
		{
			name:   "a backfilled old reading does not clear it",
			config: "no_data 20m",
			timeline: `
				0m   100
				20m  check         -> no_data started
				25m  100 at 1m     active: no_data
			`,
		},
		{
			name:   "a reading from a clock that is ahead waits from its time",
			config: "no_data 20m",
			timeline: `
				0m   100 at 10m
				25m  check
				30m  check   -> no_data started
			`,
		},
	})
}

func TestUrgentQuietsTheOther(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "a low that drops under the urgent low is quiet until it is back over it",
			config: "urgent_low 56, low 72 5m",
			timeline: `
				0m   70
				5m   70   -> low started
				10m  50   -> urgent_low started              active: urgent_low, low   alerting: urgent_low
				15m  50   -> urgent_low repeated
				20m  50   -> urgent_low repeated             // the low's repeat is due, but the urgent low says it
				25m  65   -> urgent_low cleared, low repeated   active: low   alerting: low
				30m  80   -> low cleared                     active:
			`,
		},
		{
			name:   "a drop straight under the urgent low starts the low once it is back over it",
			config: "urgent_low 56, low 72 5m",
			timeline: `
				0m   50   -> urgent_low started              active: urgent_low
				5m   50   -> urgent_low repeated             active: urgent_low, low   alerting: urgent_low
				10m  60   -> urgent_low cleared, low started   active: low
			`,
		},
		{
			name:   "a high that goes over the urgent high is quiet until it is back under it",
			config: "urgent_high 270, high 180 30m",
			timeline: `
				0m   200
				30m  280   -> urgent_high started              active: urgent_high, high   alerting: urgent_high
				35m  250   -> urgent_high cleared, high started
			`,
		},
	})
}

func TestSnooze(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "a snoozed alarm does not repeat until its snooze ends",
			config: "urgent_low 56",
			timeline: `
				0m   50     -> urgent_low started
				1m   snooze urgent_low 30m
				5m   50     active: urgent_low   alerting:
				30m  50
				31m  check  -> urgent_low repeated   alerting: urgent_low
				36m  check  -> urgent_low repeated
			`,
		},
		{
			name:   "a snoozed alarm still clears",
			config: "urgent_low 56",
			timeline: `
				0m  50  -> urgent_low started
				1m  snooze urgent_low 30m
				5m  60  -> urgent_low cleared
			`,
		},
		{
			name:   "the snooze covers it coming back on",
			config: "urgent_low 56",
			timeline: `
				0m   50     -> urgent_low started
				1m   snooze urgent_low 30m
				5m   60     -> urgent_low cleared
				10m  50     active: urgent_low   alerting:
				31m  check  -> urgent_low started
			`,
		},
		{
			name:   "it can be snoozed before it comes on",
			config: "urgent_low 56",
			timeline: `
				0m   snooze urgent_low 30m
				5m   50     active: urgent_low   alerting:
				30m  check  -> urgent_low started
			`,
		},
		{
			name:   "a snooze of 0m ends one early",
			config: "urgent_low 56",
			timeline: `
				0m  50     -> urgent_low started
				1m  snooze urgent_low 30m
				5m  check
				6m  snooze urgent_low 0m   -> urgent_low repeated
			`,
		},
		{
			name:   "snoozing one alarm leaves the others",
			config: "urgent_low 56, no_data 20m",
			timeline: `
				0m   50     -> urgent_low started
				1m   snooze urgent_low 1h
				20m  check  -> no_data started   active: urgent_low, no_data   alerting: no_data
			`,
		},
	})
}

func TestFlapping(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "going over and under it notifies no more often than staying under would",
			config: "urgent_low 56",
			timeline: `
				0m  55  -> urgent_low started
				1m  57  -> urgent_low cleared
				2m  55                          active: urgent_low   alerting: urgent_low   // alerting, but notified a minute ago
				3m  57                          active:   // it did not notify while active, so there is nothing to clear
				4m  55
				5m  55  -> urgent_low started
			`,
		},
	})
}

func TestConfigChanges(t *testing.T) {
	runTimelines(t, []timeline{
		{
			name:   "turning an alarm off while it is active clears it",
			config: "urgent_low 56",
			timeline: `
				0m  50       -> urgent_low started
				1m  config   -> urgent_low cleared   active:
			`,
		},
		{
			name:   "turning an alarm on starts it from the readings already in",
			config: "",
			timeline: `
				0m  60
				1m  config urgent_low 65   -> urgent_low started
			`,
		},
		{
			name:   "moving its threshold past the readings clears it",
			config: "urgent_low 56",
			timeline: `
				0m  55                     -> urgent_low started
				1m  config urgent_low 54   -> urgent_low cleared
			`,
		},
	})
}
