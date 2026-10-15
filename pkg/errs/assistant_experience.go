package errs

import "net/http"

var ErrAssistantRequestTooLarge = NewNormalError(NormalSubcategoryLargeLanguageModel, 50, http.StatusRequestEntityTooLarge, "assistant request body too large")

// ErrUserDataClearIncomplete never claims a partial multi-service clear completed.
var ErrUserDataClearIncomplete = NewNormalError(NormalSubcategoryLargeLanguageModel, 51, http.StatusConflict, "assistant data cleared; other data clearing result is unknown, inspect before a new clear")
