package rpc

import (
	"fmt"
	go_openrpc_reflect "github.com/etclabscore/go-openrpc-reflect"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	meta_schema "github.com/open-rpc/meta-schema"
	"go/ast"
	"net"
	"reflect"
)

func GenerateSchema(
	port string,
	receiver interface{},
) (
	schema *meta_schema.OpenrpcDocument,
	err error,
) {
	doc := &go_openrpc_reflect.Document{}
	doc.WithMeta(meta(port))
	doc.WithReflector(reflector(port))
	doc.RegisterReceiver(receiver)
	schema, err = doc.Discover()
	if err != nil {
		panic(err)
	}
	schema.Components = &meta_schema.Components{
		Schemas: (*meta_schema.SchemaComponents)(&schemas),
	}
	fixMethodResults(schema)
	return
}

func fixMethodResults(schema *meta_schema.OpenrpcDocument) {
	var methods []meta_schema.MethodObject
	for _, method := range *schema.Methods {
		if method.Result.ContentDescriptorObject.Schema == nil {
			method.Result = nil
		}
		methods = append(methods, method)
	}
	schema.Methods = (*meta_schema.Methods)(&methods)
}

var schemas = make(map[string]interface{})

func reflector(port string) go_openrpc_reflect.ReceiverRegisterer {
	var reflector = &go_openrpc_reflect.StandardReflectorT{}
	reflector.FnIsMethodEligible = isMethodEligible
	reflector.FnGetMethodParams = methodParams(reflector)
	reflector.FnGetMethodResult = methodResult(reflector)
	reflector.FnGetMethodName = methodName(port)
	reflector.FnSchemaMutations = schemaMutations
	return reflector
}

func methodName(port string) func(moduleName string, r reflect.Value, m reflect.Method, funcDecl *ast.FuncDecl) (name string, err error) {
	return func(moduleName string, r reflect.Value, m reflect.Method, funcDecl *ast.FuncDecl) (name string, err error) {
		name, err = go_openrpc_reflect.StandardReflector.GetMethodName(moduleName, r, m, funcDecl)
		if err != nil {
			return
		}
		name = port + "/" + name
		return
	}
}

func meta(port string) go_openrpc_reflect.MetaRegisterer {
	return &go_openrpc_reflect.MetaT{
		GetServersFn: func() func(listeners []net.Listener) (*meta_schema.Servers, error) {
			return func([]net.Listener) (*meta_schema.Servers, error) { return nil, nil }
		},
		GetInfoFn: func() (info *meta_schema.InfoObject) {
			title := meta_schema.InfoObjectProperties(port)
			version := meta_schema.InfoObjectVersion("0")
			return &meta_schema.InfoObject{
				Title:   &title,
				Version: &version,
			}
		},
		GetExternalDocsFn: func() (exdocs *meta_schema.ExternalDocumentationObject) {
			return nil
		},
	}
}

var isMethodEligible = func(
	method reflect.Method,
) bool {
	return true
}

func methodParams(
	registerer go_openrpc_reflect.ContentDescriptorRegisterer,
) func(
	r reflect.Value,
	m reflect.Method,
	funcDecl *ast.FuncDecl,
) (
	arr []meta_schema.ContentDescriptorObject,
	err error,
) {
	return func(
		r reflect.Value,
		m reflect.Method,
		funcDecl *ast.FuncDecl,
	) (
		arr []meta_schema.ContentDescriptorObject,
		err error,
	) {
		// A case where expanded fields arg expression would fail (if anyof `funcDecl.Type.Params` == nil)
		// should be caught by the IsMethodEligible condition.
		if funcDecl.Type.Params == nil {
			panic("unreachable")
		}

		fields := funcDecl.Type.Params.List
		for i, nf := range fields {
			ty := m.Type.In(i + 1)
			if ty.Kind() != reflect.Chan {
				cd, err := buildContentDescriptorObject(registerer, r, m, nf, ty)
				if err != nil {
					return nil, err
				}
				arr = append(arr, cd)
			}
		}
		return
	}
}

func methodResult(
	registerer go_openrpc_reflect.ContentDescriptorRegisterer,
) func(
	r reflect.Value,
	m reflect.Method,
	funcDecl *ast.FuncDecl,
) (cd meta_schema.ContentDescriptorObject, err error) {
	return func(
		r reflect.Value,
		m reflect.Method,
		funcDecl *ast.FuncDecl,
	) (
		cd meta_schema.ContentDescriptorObject,
		err error,
	) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("build content descriptor error: %w", err)
			}
		}()

		fields := funcDecl.Type.Params.List
		size := len(fields)
		field := fields[size-1]
		rootType := m.Type.In(size)
		if rootType.Kind() != reflect.Chan {
			return
		}
		replyType := rootType.Elem() // Expected chan here
		cd, err = buildContentDescriptorObject(registerer, r, m, field, replyType)
		return
	}
}

func buildContentDescriptorObject(
	registerer go_openrpc_reflect.ContentDescriptorRegisterer,
	r reflect.Value,
	m reflect.Method,
	field *ast.Field,
	ty reflect.Type,
) (
	cd meta_schema.ContentDescriptorObject,
	err error,
) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("build content descriptor error: %w", err)
		}
	}()

	name, err := registerer.GetContentDescriptorName(r, m, field)
	if err != nil {
		return cd, err
	}

	description, err := registerer.GetContentDescriptorDescription(r, m, field)
	if err != nil {
		return cd, err
	}

	summary, err := registerer.GetContentDescriptorSummary(r, m, field)
	if err != nil {
		return cd, err
	}

	required, err := registerer.GetContentDescriptorRequired(r, m, field)
	if err != nil {
		return cd, err
	}

	deprecated, err := registerer.GetContentDescriptorDeprecated(r, m, field)
	if err != nil {
		return cd, err
	}

	schema, err := registerer.GetSchema(r, m, field, ty)
	if err != nil {
		return cd, err
	}

	cd = meta_schema.ContentDescriptorObject{
		Name:        (*meta_schema.ContentDescriptorObjectName)(&name),
		Description: (*meta_schema.ContentDescriptorObjectDescription)(&description),
		Summary:     (*meta_schema.ContentDescriptorObjectSummary)(&summary),
		Schema:      &schema,
		Required:    (*meta_schema.ContentDescriptorObjectRequired)(&required),
		Deprecated:  (*meta_schema.ContentDescriptorObjectDeprecated)(&deprecated),
	}
	return
}

var schemaMutations = func(ty reflect.Type) []func(*spec.Schema) func(*spec.Schema) error {
	pkgPath := ty.PkgPath()
	if pkgPath == "" {
		switch ty.Kind() {
		case
			reflect.Chan,
			reflect.Map,
			reflect.Array,
			reflect.Slice:
			pkgPath = ty.Elem().PkgPath()
		}
	}
	return []func(*spec.Schema) func(*spec.Schema) error{
		func(rootSchema *spec.Schema) func(*spec.Schema) error {
			return func(mutSchema *spec.Schema) error {
				for id, defSchema := range mutSchema.Definitions {
					schemas[pkgPath+"/"+id] = defSchema
					for name, prop := range defSchema.Properties {
						err := fixReference(pkgPath, &prop)
						if err != nil {
							return err
						}
						defSchema.Properties[name] = prop
					}
				}
				mutSchema.Definitions = nil
				err := fixReference(pkgPath, mutSchema)
				return err
			}
		},
	}
}

func fixReference(pkgPath string, s *spec.Schema) error {
	var schema = s
	if s.Items != nil {
		schema = s.Items.Schema
	}
	tokens := schema.Ref.GetPointer().DecodedTokens()
	if len(tokens) == 0 {
		return nil
	}
	ref := pkgPath + "/" + tokens[1]
	jsonRef, err := jsonreference.New(ref)
	if err != nil {
		return err
	}
	schema.Ref = spec.Ref{Ref: jsonRef}
	return nil
}
