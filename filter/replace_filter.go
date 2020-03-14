package filter

import (
	"regexp"
	"strings"

	"github.com/childe/gohangout/field_setter"
	"github.com/childe/gohangout/value_render"
	"github.com/golang/glog"
)

type replaceConfig struct {
	s   field_setter.FieldSetter
	v   value_render.ValueRender
	old string
	new string
	n   int
}

type ReplaceFilter struct {
	config map[interface{}]interface{}
	fields []replaceConfig
}

func (l *MethodLibrary) NewReplaceFilter(config map[interface{}]interface{}) *ReplaceFilter {
	p := &ReplaceFilter{
		config: config,
		fields: make([]replaceConfig, 0),
	}

	if fieldsI, ok := config["fields"]; ok {

		for fieldI, configI := range fieldsI.(map[interface{}]interface{}) {
			fieldSetter := field_setter.NewFieldSetter(fieldI.(string))
			if fieldSetter == nil {
				glog.Fatalf("could build field setter from %s", fieldI.(string))
			}

			v := value_render.GetValueRender2(fieldI.(string))

			rConfig := configI.([]interface{})
			if len(rConfig) == 2 {
				t := replaceConfig{
					fieldSetter,
					v,
					rConfig[0].(string),
					rConfig[1].(string),
					-1,
				}
				p.fields = append(p.fields, t)
			} else if len(rConfig) == 3 {
				t := replaceConfig{
					fieldSetter,
					v,
					rConfig[0].(string),
					rConfig[1].(string),
					rConfig[2].(int),
				}
				p.fields = append(p.fields, t)
			} else {
				glog.Fatal("invalid fields config in replace filter")
			}
		}
	} else {
		glog.Fatal("fileds must be set in replace filter plugin")
	}

	return p
}

// if the filed is not string, return false, else true
func (p *ReplaceFilter) Filter(event map[string]interface{}) (map[string]interface{}, bool) {
	success := true
	for _, f := range p.fields {
		value := f.v.Render(event)
		if value == nil {
			continue
		}
		if s, ok := value.(string); ok {
			var new string
			if strings.HasPrefix(f.old, "sensitive-mobile-") {
				regexString := strings.Replace(f.old, "sensitive-mobile-", "", -1)
				rege, _ := regexp.Compile(regexString)
				new = rege.ReplaceAllStringFunc(s, replaceMobileNumber)
			} else if strings.HasPrefix(f.old, "sensitive-email-") {
				regexString := strings.Replace(f.old, "sensitive-email-", "", -1)
				rege, _ := regexp.Compile(regexString)
				new = rege.ReplaceAllStringFunc(s, replaceEmailNumber)
			} else {
				new = strings.Replace(s, f.old, f.new, f.n)
			}
			f.s.SetField(event, new, "", true)
		} else {
			success = false
		}
	}
	return event, success

}

func replaceMobileNumber(mobile string) string {
	n := len(mobile)
	result := make([]byte, n)

	for i := 0; i < n; i++ {
		if i < 7 {
			result[i] = '*'
		} else {
			result[i] = mobile[i]
		}
	}
	return string(result)
	//return "[****** sensitive ******]"
}

func replaceEmailNumber(email string) string {
	n := len(email)
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		if i > 1 && i < n-4 {
			result[i] = '*'
		} else {
			result[i] = email[i]
		}
	}
	return string(result)
}
