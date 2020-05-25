package roamer

// VerifyNoDirty checks that the environment has no dirty migrations.
func (e *Environment) VerifyNoDirty() (bool, error) {
	appliedMigrations, err := e.ListAppliedMigrations()
	if err != nil {
		return false, err
	}

	for _, appliedMigration := range appliedMigrations {
		if appliedMigration.Dirty {
			return false, nil
		}
	}

	return true, nil
}

// VerifyExist checks that that all applied migrations exist on disk.
func (e *Environment) VerifyExist() (bool, error) {
	appliedMigrations, err := e.ListAppliedMigrations()
	if err != nil {
		return false, err
	}

	for _, appliedMigration := range appliedMigrations {
		_, err := e.GetMigrationByID(appliedMigration.ID)
		if err != nil {
			if err == ErrMigrationNotFound {
				return false, nil
			}

			return false, err
		}
	}

	return true, nil
}

// VerifyOrder checks that the order of migrations on disk matches the order in the history.
func (e *Environment) VerifyOrder() (bool, error) {
	allMigrations, err := e.ListAllMigrations()
	if err != nil {
		return false, err
	}

	appliedMigrations, err := e.ListAppliedMigrations()
	if err != nil {
		return false, err
	}

	if len(appliedMigrations) > len(allMigrations) {
		return false, nil
	}

	for i, appliedMigration := range appliedMigrations {
		if allMigrations[i].ID != appliedMigration.ID {
			return false, nil
		}
	}

	return true, nil
}

// VerifySafeToApply checks that it is safe to apply migrations, running all other verification checks.
func (e *Environment) VerifySafeToApply() (bool, error) {
	noDirty, err := e.VerifyNoDirty()
	if err != nil {
		return false, err
	}
	if !noDirty {
		return false, nil
	}

	exist, err := e.VerifyExist()
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}

	order, err := e.VerifyOrder()
	if err != nil {
		return false, err
	}
	if !order {
		return false, nil
	}

	return true, nil
}
