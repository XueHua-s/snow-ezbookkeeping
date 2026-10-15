package templates

import (
	"sync"
	"testing"
)

func TestConcurrentTemplateLoad(t *testing.T) {
	t.Chdir("../..")
	var workers sync.WaitGroup
	for i := 0; i < 32; i++ {
		workers.Go(func() {
			for _, name := range []KnownTemplate{SYSTEM_PROMPT_PERSONAL_FINANCE_ASSISTANT, SYSTEM_PROMPT_RECEIPT_IMAGE_RECOGNITION} {
				if _, err := GetTemplate(name); err != nil {
					t.Errorf("load %s: %v", name, err)
				}
			}
		})
	}
	workers.Wait()
}
