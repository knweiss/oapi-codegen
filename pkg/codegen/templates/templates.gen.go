package templates

import "text/template"

var templates = map[string]string{"imports.tmpl": `// This is an autogenerated file, any edits which you make here will be lost!
package {{.PackageName}}

import (
{{range .Imports}} "{{.}}"
{{end}})
`,
	"parameters.tmpl": `{{range .Types}}
// Type definition for component parameter "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"register.tmpl": `func RegisterHandlers(router codegen.EchoRouter, si ServerInterface) {
    wrapper := ServerInterfaceWrapper{
        Handler: si,
    }
{{range .}}router.{{.Method}}("{{.Path | swaggerUriToEchoUri}}", wrapper.{{.OperationId}})
{{end}}
}`,
	"request-bodies.tmpl": `{{range .Types}}
// Type definition for component requestBodies "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"responses.tmpl": `{{range .Types}}
// Type definition for component response "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"schemas.tmpl": `{{range .Types}}
// Type definition for component schema "{{.JsonTypeName}}"
type {{.TypeName}} {{.TypeDef}}
{{end}}
`,
	"server-interface.tmpl": `{{range .}}

{{if .Params}}
// Parameters object for {{.OperationId}}
type {{.OperationId}}Params struct {
{{range .Params}}
    {{.GoName}} {{if not .Required}}*{{end}}{{.TypeDef}} {{.JsonTag}}{{end}}
}
{{end}}

{{if .HasBody}}
{{if .GetBodyDefinition.CustomType}}
// Request body for {{.OperationId}} for application/json ContentType
type {{.OperationId}}RequestBody {{.GetBodyDefinition.TypeDef}}
{{end}}
{{end}}

{{end}}

type ServerInterface interface {
{{range .}}// {{.Summary}} ({{.Method}} {{.Path}})
{{.OperationId}}(ctx echo.Context{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params {{.OperationId}}Params{{end}}) error
{{end}}
}
`,
	"wrappers.tmpl": `type ServerInterfaceWrapper struct {
    Handler ServerInterface
}

{{range .}}// Wrapper for {{.OperationId}}
func (w *ServerInterfaceWrapper) {{.OperationId}} (ctx echo.Context) error {
    var err error
{{range .PathParams}}// ------------- Path parameter "{{.ParamName}}" -------------
    var {{.GoName}} {{.TypeDef}}
    err = codegen.BindStringToObject(ctx.Param("{{.ParamName}}"), &{{.GoName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
{{end}}

{{if .RequiresParamObject}}
    // Parameter object where we will unmarshal all parameters from the
    // context.
    var params {{.OperationId}}Params
{{range .QueryParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} query parameter "{{.ParamName}}" -------------
var {{.GoName}} {{.TypeDef}}
{{if .Required}}
    err = codegen.BindStringToObject(ctx.QueryParam("{{.ParamName}}"), &{{.GoName}})
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
    }
    params.{{.GoName}} = {{.GoName}}
{{else}}
    if ctx.QueryParam("{{.ParamName}}") != "" {
        err = codegen.BindStringToObject(ctx.QueryParam("{{.ParamName}}"), &{{.GoName}})
        if err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
        }
        params.{{.GoName}} = &{{.GoName}}
    }{{end}}
{{end}}
{{if .HeaderParams}}
    headers := ctx.Request().Header
{{range .HeaderParams}}// ------------- {{if .Required}}Required{{else}}Optional{{end}} header parameter "{{.ParamName}}" -------------
    var {{.GoName}} {{.TypeDef}}
    {
        valueList, found := headers["{{.ParamName}}"]
{{if .Required}}
        if !found {
            // This should never happen, as Swagger would catch it during
            // validation, but just in case...
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter {{.ParamName}} is required, but not found"))
        }
{{end}}
        n := len(valueList)
        if n != 1 {
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for {{.ParamName}}, got %d", n))
        }
{{if .Required}}
        err = codegen.BindStringToObject(ctx.Param("{{.ParamName}}"), &{{.GoName}})
        if err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
        }
        params.{{.GoName}} = {{.GoName}}
{{else}}
        if found {
            err = codegen.BindStringToObject(ctx.Param("{{.ParamName}}"), &{{.GoName}})
            if err != nil {
                return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter {{.ParamName}}: %s", err))
            }
            params.{{.GoName}} = &{{.GoName}}
        }
{{end}}

    }
{{end}}
{{end}}
{{end}}
    // Invoke the callback with all the unmarshalled arguments
    err = w.Handler.{{.OperationId}}(ctx{{genParamNames .PathParams}}{{if .RequiresParamObject}}, params{{end}})
    return err
}
{{end}}`,
}

// Parse parses declared templates.
func Parse(t *template.Template) (*template.Template, error) {
	for name, s := range templates {
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		if _, err := tmpl.Parse(s); err != nil {
			return nil, err
		}
	}
	return t, nil
}

