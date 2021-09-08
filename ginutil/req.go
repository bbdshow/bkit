package ginutil

import (
	"fmt"
	"github.com/bbdshow/bkit/errc"
	"github.com/bbdshow/bkit/logs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net"
	"reflect"
	"strings"
	"sync"

	zhongwen "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zh_trans "github.com/go-playground/validator/v10/translations/zh"
)

func validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i, ipStr := range items {
		ipStr = strings.TrimSpace(ipStr)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return "", false
		}

		// We need to return the first IP in the list, but,
		// we should not early return since we need to validate that
		// the rest of the header is syntactically valid
		if i == 0 {
			clientIP = ipStr
			valid = true
		}
	}
	return
}

// ClientIP Get X-Forwarded-For, X-Real-Ip
func ClientIP(c *gin.Context) string {
	remoteIPHeaders := []string{"X-Forwarded-For", "X-Real-Ip"}
	for _, headerName := range remoteIPHeaders {
		ip, valid := validateHeader(c.GetHeader(headerName))
		if valid {
			return ip
		}
	}
	remoteIP, _ := c.RemoteIP()
	if remoteIP == nil {
		return ""
	}
	return remoteIP.String()
}

func ShouldBind(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		logs.Qezap.Warn("ParamException", zap.Any("RequestParam", obj), zap.Any("Exception", err), logs.Qezap.FieldTraceID(c.Request.Context()))
		if !gin.IsDebugging() {
			// hide specific param info
			return errc.ErrParamInvalid
		}
		return errc.ErrParamInvalid.MultiErr(err)
	}
	return nil
}

func ValidateStruct(obj interface{}) error {
	return binding.Validator.ValidateStruct(obj)
}

var stdValidator = validator.New()
var stdZhTranslation = zhTranslation()

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			if errs, ok := err.(validator.ValidationErrors); ok {
				return fmt.Errorf("%v", errs.Translate(stdZhTranslation))
			}
			return err
		}
	}

	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = stdValidator
		v.validate.SetTagName("binding")
		// add any custom validations etc. here
	})
}

func kindOfData(data interface{}) reflect.Kind {

	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func zhTranslation() ut.Translator {
	zh := zhongwen.New()
	uni := ut.New(zh, zh)
	trans, _ := uni.GetTranslator("zh")
	if err := zh_trans.RegisterDefaultTranslations(stdValidator, trans); err != nil {
		fmt.Println("RegisterDefaultTranslations", err.Error())
	}
	return trans
}
