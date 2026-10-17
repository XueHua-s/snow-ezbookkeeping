package services

// categoryUsage computes metadata activity in the same UID-bound read snapshot.
// One bounded full-ledger scan uses the existing logical-transfer normalization;
// name/type filtering and entity pagination apply only after counts are complete.
// This is record usage, not signed amount or activity within a requested period.
func (e *assistantQueryExecution) categoryUsage() (map[int64]int64, map[int64]int64, error) {
	direct := map[int64]int64{}
	if err := e.ScanTransactions(nil, func(tx *assistantLogicalTransaction) error {
		if tx.Category != nil {
			direct[tx.Category.CategoryId]++
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	children := map[int64][]int64{}
	for id, category := range e.Metadata.Categories {
		children[category.ParentCategoryId] = append(children[category.ParentCategoryId], id)
	}
	subtree := map[int64]int64{}
	visiting := map[int64]bool{}
	var count func(int64) (int64, error)
	count = func(id int64) (int64, error) {
		if err := e.Context.Err(); err != nil {
			return 0, err
		}
		if visiting[id] {
			return 0, assistantInconsistent()
		}
		if total, ok := subtree[id]; ok {
			return total, nil
		}
		visiting[id] = true
		total := direct[id]
		for _, child := range children[id] {
			n, err := count(child)
			if err != nil {
				return 0, err
			}
			total += n
		}
		delete(visiting, id)
		subtree[id] = total
		return total, nil
	}
	for id := range e.Metadata.Categories {
		if _, err := count(id); err != nil {
			return nil, nil, err
		}
	}
	return direct, subtree, nil
}
