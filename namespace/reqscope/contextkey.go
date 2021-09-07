package reqscope

import (
	"context"

	"golang.org/x/text/language"
)

type contextKey int

const (
	keyLanguage contextKey = iota
)

func Language(ctx context.Context) language.Tag {
	if lang, ok := ctx.Value(keyLanguage).(language.Tag); ok {
		return lang
	}
	return language.English
}

func SetLanguage(ctx context.Context, lang language.Tag) context.Context {
	return context.WithValue(ctx, keyLanguage, lang)
}
