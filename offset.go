package roamer

import (
	"strconv"
)

// ResolveIDOrOffset looks up and returns the requested migration.
// It handles absolute IDs, absolute offsets (such as @2), and relative offsets (such as @+1).
func (e *Environment) ResolveIDOrOffset(idOrOffset string) (*Migration, error) {
	if len(idOrOffset) == 0 {
		return nil, InvalidInputError{idOrOffset}
	}

	allMigrations, err := e.ListAllMigrations()
	if err != nil {
		return nil, err
	}

	if idOrOffset[0] == '@' {
		// it's an offset
		offsetDetails := idOrOffset[1:]
		if len(offsetDetails) == 0 {
			return nil, InvalidInputError{idOrOffset}
		}

		// try to parse the index
		requestedIndex, err := strconv.Atoi(offsetDetails)
		if err != nil {
			return nil, InvalidInputError{idOrOffset}
		}

		if offsetDetails[0] == '+' || offsetDetails[0] == '-' {
			// it's a relative offset
			lastApplied, err := e.GetLastAppliedMigration()
			if err != nil {
				return nil, err
			}

			// find the current index
			lastAppliedIndex := -1
			if lastApplied != nil {
				for i, migration := range allMigrations {
					if migration.ID == lastApplied.ID {
						lastAppliedIndex = i
						break
					}
				}

				if lastAppliedIndex == -1 {
					return nil, ErrMigrationNotFound
				}
			}

			requestedIndex = (lastAppliedIndex + 1) + requestedIndex
		}

		if len(allMigrations) < requestedIndex || requestedIndex < 0 {
			return nil, OffsetBoundError{idOrOffset}
		}

		if requestedIndex == 0 {
			return nil, nil
		}

		return &allMigrations[requestedIndex-1], nil
	}

	// it must be an id
	migration, err := e.GetMigrationByID(idOrOffset)
	if err != nil {
		return nil, err
	}

	return &migration, err
}
