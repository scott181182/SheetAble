package controllers

import (
	"time"

	"github.com/SheetAble/SheetAble/backend/api/models"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/handler"

	"github.com/gin-gonic/gin"
)



func (server *Server) ResolveSheets(p graphql.ResolveParams) (interface{}, error) {
	page, _ := p.Args["page"].(int)
	pageSize, _ := p.Args["pageSize"].(int)

	if pageSize == 0 { pageSize = 50 }
	if page == 0 { page = 1 }

	pagination := models.Pagination{
		Sort:  "updated_at desc",
		Limit: pageSize,
		Page:  page,
	}

	var sheet models.Sheet
	pageNew, err := sheet.List(server.DB, pagination, "")
	return pageNew.Rows, err
}
func (server *Server) ResolveComposers(p graphql.ResolveParams) (interface{}, error) {
	page, _ := p.Args["page"].(int)
	pageSize, _ := p.Args["pageSize"].(int)

	if pageSize == 0 { pageSize = 50 }
	if page == 0 { page = 1 }

	pagination := models.Pagination{
		Sort:  "updated_at desc",
		Limit: pageSize,
		Page:  page,
	}

	var composer models.Composer
	pageNew, err := composer.List(server.DB, pagination)
	return pageNew.Rows, err
}

func (server *Server) GetGraphQLSchema() *graphql.Schema {
	datetimeType := graphql.NewScalar(graphql.ScalarConfig{
		Name: "DateTime",
		Description: "An ISO-formatted DateTime String",
		Serialize: func(value interface{}) interface{} {
			dt, isTime := value.(time.Time)
			if !isTime {
				// Not sure how to return errors here
				// return graphql.Errorf("Cannot serialize non-time.Time as DateTime")
				return nil
			}
			return dt.Format(time.RFC3339)
		},
		ParseValue: func(value interface{}) interface{} {
			str, isString := value.(string)
			if !isString {
				// Not sure how to return errors here
				// return nil, graphql.Errorf("Cannot parse non-string literal as DateTime")
				return nil
			}
			dt, err := time.Parse(time.RFC3339, str)
			if err != nil {
				// Not sure how to return errors here
				// return nil, graphql.Errorf("Cannot parse non-string literal as DateTime")
				return nil
			}
			return dt
		},
		ParseLiteral: func(valueAST ast.Value) interface{} {
			// Could also check `valueAST.GetKind()` here
			str, isString := valueAST.GetValue().(string)
			if !isString {
				// Not sure how to return errors here
				// return nil, graphql.Errorf("Cannot parse non-string literal as DateTime")
				return nil
			}
			dt, err := time.Parse(time.RFC3339, str)
			if err != nil {
				// Not sure how to return errors here
				// return nil, graphql.Errorf("Cannot parse non-string literal as DateTime")
				return nil
			}
			return dt
		},
	})

	sheetType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Sheet",
		Fields: graphql.Fields{
			"SheetName": &graphql.Field{ Type: graphql.String },
			"SafeSheetName": &graphql.Field{ Type: graphql.String, },
			"ComposerName": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					src, _ := p.Source.(*models.Sheet)
					return src.Composer, nil
				},
			},
			"SafeComposerName": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					src, _ := p.Source.(*models.Sheet)
					return src.SafeComposer, nil
				},
			},
			"InformationText": &graphql.Field{ Type: graphql.String, },
			"PdfUrl": &graphql.Field{ Type: graphql.String, },
			"Tags": &graphql.Field{ Type: graphql.NewList(graphql.String), },

			"CreatedAt": &graphql.Field{ Type: datetimeType, },
			"UpdatedAt": &graphql.Field{ Type: datetimeType, },
			"ReleaseDate": &graphql.Field{ Type: datetimeType, },
			"UploaderID": &graphql.Field{ Type: graphql.Int, },
		},
	})
	composerType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Composer",
		Fields: graphql.Fields{
			"SafeName": &graphql.Field{ Type: graphql.String },
			"Name": &graphql.Field{ Type: graphql.String, },
			"PortraitURL": &graphql.Field{ Type: graphql.String, },
			"Epoch": &graphql.Field{ Type: graphql.String, },

			"CreatedAt": &graphql.Field{ Type: datetimeType, },
			"UpdatedAt": &graphql.Field{ Type: datetimeType, },
		},
	})

	sheetType.AddFieldConfig("Composer", &graphql.Field{
		Type: composerType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			src, _ := p.Source.(*models.Sheet)
			var composerModel models.Composer
			composer, err := composerModel.FindComposerBySafeName(server.DB, src.SafeComposer)
			return composer, err
		},
	})
	composerType.AddFieldConfig("Sheets", &graphql.Field{
		Type: graphql.NewList(sheetType),
		Args: graphql.FieldConfigArgument{
			"page": &graphql.ArgumentConfig{ Type: graphql.Int },
			"pageSize": &graphql.ArgumentConfig{ Type: graphql.Int },
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			src, _ := p.Source.(*models.Composer)
			page, _ := p.Args["page"].(int)
			pageSize, _ := p.Args["pageSize"].(int)
		
			if pageSize == 0 { pageSize = 50 }
			if page == 0 { page = 1 }
		
			pagination := models.Pagination{
				Sort:  "updated_at desc",
				Limit: pageSize,
				Page:  page,
			}

			var sheet models.Sheet
			pageNew, err := sheet.List(server.DB, pagination, src.SafeName)
			return pageNew.Rows, err
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"sheets": &graphql.Field{
				Type: graphql.NewList(sheetType),
				Description: "List of Sheets",
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{ Type: graphql.Int },
					"pageSize": &graphql.ArgumentConfig{ Type: graphql.Int },
				},
				Resolve: server.ResolveSheets,
			},
			"composers": &graphql.Field{
				Type: graphql.NewList(composerType),
				Description: "List of Composers",
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{ Type: graphql.Int },
					"pageSize": &graphql.ArgumentConfig{ Type: graphql.Int },
				},
				Resolve: server.ResolveComposers,
			},
		},
	})
	
	SheetableSchema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	return &SheetableSchema
}
  

func (server *Server) MakeGraphQLHandler() gin.HandlerFunc {
	h := handler.New(&handler.Config{
		Schema: server.GetGraphQLSchema(),
		Pretty: true,
		GraphiQL: false,
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
