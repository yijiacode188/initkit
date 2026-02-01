package lib

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var Bundle *i18n.Bundle

type i18nConf struct {
	Path string `mapstructure:"path"`
}

func initI18n() error {
	conf := &i18nConf{}

	if !IsSetConf("i18n") {
		return errors.New("未配置i18n")
	}
	err := viperConf.UnmarshalKey("i18n", &conf)
	if err != nil {
		return err
	}
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	err, list := ListFiles(conf.Path)
	if err != nil {
		return err
	}
	for _, item := range list {
		Bundle.MustLoadMessageFile(item)
	}
	return nil
}
func GetI18Message(lang string, key string, data map[string]interface{}, pluralCount interface{}) string {
	if lang == "" {
		lang = "en"
	}
	localized := i18n.NewLocalizer(Bundle, lang)
	str, err := localized.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		PluralCount:  pluralCount,
		TemplateData: data,
	})
	if err != nil {
		return "Error"
	}
	return str
}
func GetI18nStr(lang, key string) string {
	return GetI18Message(lang, key, map[string]interface{}{}, 1)
}
