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

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// SetTPAccountActions creates a new AccountActions profile within a tariff plan
func (apiv1 *APIerSv1) SetTPAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "LoadId", "Tenant", "Account", "ActionPlanId"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apiv1.StorDb.SetTPAccountActions([]*utils.TPAccountActions{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPAccountActionsByLoadId struct {
	TPid   string // Tariff plan id
	LoadId string // AccountActions id
}

// GetTPAccountActionsByLoadId queries specific AccountActions profile on tariff plan
func (apiv1 *APIerSv1) GetTPAccountActionsByLoadId(attrs utils.TPAccountActions, reply *[]*utils.TPAccountActions) error {
	mndtryFlds := []string{"TPid", "LoadId"}
	if len(attrs.Account) != 0 { // If account provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, "Tenant", "Account")
	}
	if missing := utils.MissingStructFields(&attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aas, err := apiv1.StorDb.GetTPAccountActions(&attrs)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = aas
	return nil
}

type AttrGetTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // DerivedCharge id
}

// GetTPAccountActions queries specific DerivedCharge on tariff plan
func (apiv1 *APIerSv1) GetTPAccountActions(attrs AttrGetTPAccountActions, reply *utils.TPAccountActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AccountActionsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter := &utils.TPAccountActions{TPid: attrs.TPid}
	if err := filter.SetAccountActionsId(attrs.AccountActionsId); err != nil {
		return err
	}
	aas, err := apiv1.StorDb.GetTPAccountActions(filter)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *aas[0]
	return nil
}

type AttrGetTPAccountActionIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPAccountActionLoadIds queries AccountActions identities on specific tariff plan.
func (apiv1 *APIerSv1) GetTPAccountActionLoadIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apiv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAccountActions,
		utils.TPDistinctIds{"loadid"}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// GetTPAccountActionIds queries DerivedCharges identities on specific tariff plan.
func (apiv1 *APIerSv1) GetTPAccountActionIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apiv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAccountActions,
		utils.TPDistinctIds{"loadid", "tenant", "account"}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPAccountActions removes specific AccountActions on Tariff plan
func (apiv1 *APIerSv1) RemoveTPAccountActions(attrs AttrGetTPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Account"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aa := engine.TpAccountAction{Tpid: attrs.TPid}
	if err := aa.SetAccountActionId(attrs.AccountActionsId); err != nil {
		return err
	}
	if err := apiv1.StorDb.RemTpData(utils.TBLTPAccountActions, aa.Tpid,
		map[string]string{"loadid": aa.Loadid, "tenant": aa.Tenant, "account": aa.Account}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
