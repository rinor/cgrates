//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRatingSubjectSetDefault(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true,
	"balance_rating_subject":{
		"*data":"",
	}
},

"cdrs": {
	"enabled": true,
	"rals_conns": ["*localhost"]
},

"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}

}`

	client, _, shutdown, err := setupTest(t, "TestBalanceBlocker", utils.EmptyString, utils.EmptyString, content, map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP_DATA,*asap,10
PACKAGE_1001,ACT_TOPUP_MON,*asap,10
`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_DATA,*topup_reset,,,data1,*data,,*any,,,*unlimited,,102400,10,false,false,10
ACT_TOPUP_MON,*topup_reset,,,money1,*monetary,,*any,,,*unlimited,,250,10,false,false,10
`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_DATA,*any,RT_DATA,*up,0,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_DATA,0,1,1024,1024,0`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_DATA,DR_DATA,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,data,1001,2022-01-14T00:00:00Z,RP_DATA,`,
	})

	if err != nil {
		t.Fatalf("failed to do setup for test: %v", err)
	}

	defer shutdown()

	t.Run("CheckInitialBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		} else if len(acnt.BalanceMap) != 2 {
			t.Fatalf("expected account to have only one balance of type *data, received %v", utils.ToJSON(acnt))
		} else if balanceD := acnt.BalanceMap[utils.MetaData][0]; balanceD.ID != "data1" || balanceD.Value != 102400 {
			t.Fatalf("received account with unexpected balance: %v", balanceD)
		} else if balanceM := acnt.BalanceMap[utils.MetaMonetary][0]; balanceM.ID != "money1" || balanceM.Value != 250 {
			t.Fatalf("received account with unexpected balance: %v", balanceM)
		}
	})

	t.Run("ProcessCDR", func(t *testing.T) {
		var reply []*utils.EventWithFlags
		if err := client.Call(context.Background(), utils.CDRsV2ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "data",
						utils.ToR:          utils.MetaData,
						utils.OriginID:     "processCDR",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaPrepaid,
						utils.AccountField: "1001",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2022, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2022, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        10000,
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		} else if ev := reply[0].Event; ev[utils.Cost] != 10.0 {
			t.Fatalf("Expected Cost to be 5,received %v", ev[utils.Cost])
		}
	})

	t.Run("CheckFinalBalance", func(t *testing.T) {
		var acnt engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
			t.Fatal(err)
		} else if len(acnt.BalanceMap) != 2 {
			t.Fatalf("expected account to have 2 balances  *monetary and *data, received %v", acnt)
		} else if balanceM := acnt.BalanceMap[utils.MetaMonetary][0]; balanceM.ID != "money1" || balanceM.Value != 240.0 {
			t.Fatalf("received account with unexpected balance: %v", balanceM)
		} else if balanceD := acnt.BalanceMap[utils.MetaData][0]; balanceD.ID != "data1" || balanceD.Value != 92160 {
			t.Fatalf("received account with unexpected balance: %v", balanceD)
		}
	})
}