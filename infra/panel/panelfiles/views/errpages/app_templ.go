// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778
package errpages

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "github.com/CSPF-Founder/iva/panel/internal/sessions"

func AppErrPage(flashes []sessions.SessionFlash) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>Error Occurred!</title><link rel=\"stylesheet\" href=\"/static/vendor/simplebar/css/simplebar.css\"><link rel=\"stylesheet\" href=\"/static/css/vendors/simplebar.css\"><!-- Main styles for this application--><link href=\"/static/css/style.min.css\" rel=\"stylesheet\"></head><body class=\"nav-md\"><div class=\"bg-light min-vh-100 d-flex flex-row align-items-center dark:bg-transparent\"><div class=\"container\"><div class=\"row justify-content-center\"><div class=\"col-md-6\"><div class=\"clearfix\"><h1 class=\"display-3 me-4\">Error Occurred !</h1>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if len(flashes) > 0 {
			for _, item := range flashes {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h4 class=\"pt-3\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var2 string
				templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(item.Message)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/errpages/app.templ`, Line: 27, Col: 41}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h4>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
		} else {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h4 class=\"pt-3\">Sorry, an unknown error has occurred - Please contact the admin</h4>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h4 class=\"pt-3\">If the problem persists feel free to contact us</h4></div></div></div></div></div></body><script src=\"/static/vendor/@coreui/coreui-pro/js/coreui.bundle.min.js\"></script><script src=\"/static/vendor/simplebar/js/simplebar.min.js\"></script></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
