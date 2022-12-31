package controllers

import (
	"github.com/SheetAble/SheetAble/backend/api/models"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"

	"github.com/gin-gonic/gin"
)



var sheetType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Sheet",
	Fields: graphql.Fields{
		"SheetName": &graphql.Field{ Type: graphql.String },
		"SafeSheetName": &graphql.Field{ Type: graphql.String, },
		"Composer": &graphql.Field{ Type: graphql.String, },
		"SafeComposer": &graphql.Field{ Type: graphql.String, },
		"InformationText": &graphql.Field{ Type: graphql.String, },
		"PdfUrl": &graphql.Field{ Type: graphql.String, },
		"Tags": &graphql.Field{ Type: graphql.NewList(graphql.String), },

		"CreatedAt": &graphql.Field{ Type: graphql.String, },
		"UpdatedAt": &graphql.Field{ Type: graphql.String, },
		"ReleaseDate": &graphql.Field{ Type: graphql.String, },
		"UploaderID": &graphql.Field{ Type: graphql.Int, },
	},
})

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

func (server *Server) GetGraphQLSchema() *graphql.Schema {
	var rootQuery = graphql.NewObject(graphql.ObjectConfig{
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
			} ,
		},
	})
	
	var SheetableSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
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
