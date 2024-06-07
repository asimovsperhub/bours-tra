package translators

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var Trans ut.Translator

func InitTranslators(locale string) (err error) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New() //中文翻译器
		enT := en.New() //英文翻译器

		uni := ut.New(zhT, zhT, enT)

		var ok bool

		Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uin.GetTranslator(%s) failed", locale)
		}

		switch locale {
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, Trans)
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, Trans)
		default:
			err = zhTranslations.RegisterDefaultTranslations(v, Trans)
		}
		return
	}
	return
}
