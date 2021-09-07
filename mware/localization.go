package mware

import (
	"errors"
	"net/http"

	"github.com/stevealexrs/Go-Libra/namespace/cookiens"
	"github.com/stevealexrs/Go-Libra/namespace/reqscope"
	"golang.org/x/text/language"
)

func Localization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept-Language")
		langCookie, err := r.Cookie(cookiens.Language)
		var override string
		if errors.Is(err, http.ErrNoCookie) {
			override = ""
		} else {
			override = langCookie.Value
		}

		matcher := language.NewMatcher([]language.Tag{
			language.English,
			language.Chinese,
			language.Malay,
		})

		tag, _ := language.MatchStrings(matcher, override, accept)
		ctx := reqscope.SetLanguage(r.Context(), tag)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}