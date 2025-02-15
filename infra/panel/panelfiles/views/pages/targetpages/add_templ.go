// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778
package targetpages

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "github.com/CSPF-Founder/iva/panel/views/helpers"

func AddTarget(title string, isDS bool) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"row pt-3\"><div class=\"col-lg-12 col-sm-12 pb-3 fpt14\"><h3 class=\"font-weight-bold\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/pages/targetpages/add.templ`, Line: 8, Col: 39}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h3></div></div><div class=\"row pr-3\"><div class=\"col-12 p-0 \"><div class=\"scroller card mb-4 col-sm-9 col-lg-12 table-wrapper ml-0 p-0 leftPadding-10\"><div class=\"direction-r pr-3 card-body\"><div class=\"fs-5 text-center\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/pages/targetpages/add.templ`, Line: 15, Col: 42}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div><hr><div class=\"tab-content rounded-bottom\"><div class=\"tab-pane p-3 active preview\" role=\"tabpanel\"><form action=\"/targets/add\" id=\"add-scan-form\" method=\"POST\" enctype=\"multipart/form-data\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if isDS {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<input type=\"hidden\" name=\"is_ds\" value=\"true\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"row\"><div class=\"col-lg-2 col-sm-0\"></div><div class=\"col-lg-1 col-sm-1\"><label style=\"margin-top:5px;\">Target :</label></div><div class=\"col-lg-7 col-sm-7\"><input class=\"form-control sharpedge\" type=\"text\" name=\"target_address\" placeholder=\"Eg: 192.168.0.1, 192.168.0.1/24, http://app.dev.example/\" required></div><div class=\"col-lg-2 col-sm-2\"></div></div><div class=\"row pt-4\"><div class=\"col-lg-3 col-sm-3\"></div><div class=\"col-lg-4 col-sm-4\"><div class=\" \"><button id=\"add-scan-btn\" type=\"submit\" class=\"btn btn-primary center\"><b>Add</b></button></div></div><div class=\"col-lg-4 col-sm-4\"></div></div></form></div></div></div></div></div></div><script type=\"module\" src=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var4 string
		templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(helpers.AssetPath("app/scans.js"))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/pages/targetpages/add.templ`, Line: 59, Col: 62}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"></script>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
