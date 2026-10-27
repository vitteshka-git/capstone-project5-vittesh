package graphqlapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"

	"order-management/model"
	"order-management/service"
)

func NewSchema(svc *service.OrderService) (graphql.Schema, error) {
	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.Int},
			"customerId":  &graphql.Field{Type: graphql.Int},
			"productId":   &graphql.Field{Type: graphql.Int},
			"quantity":    &graphql.Field{Type: graphql.Int},
			"unitPrice":   &graphql.Field{Type: graphql.Float},
			"totalAmount": &graphql.Field{Type: graphql.Float},
			"status":      &graphql.Field{Type: graphql.String},
			"createdAt":   &graphql.Field{Type: graphql.DateTime},
			"updatedAt":   &graphql.Field{Type: graphql.DateTime},
		},
	})

	createOrderInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateOrderInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"customerId": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"productId": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"quantity": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"order": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if !ok || id <= 0 {
						return nil, nil
					}
					return svc.GetOrder(p.Context, int64(id))
				},
			},
			"orders": &graphql.Field{
				Type: graphql.NewList(orderType),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 20,
					},
					"offset": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 0,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					limit, _ := p.Args["limit"].(int)
					offset, _ := p.Args["offset"].(int)

					if limit <= 0 {
						limit = 20
					}
					if limit > 100 {
						limit = 100
					}
					if offset < 0 {
						offset = 0
					}

					return svc.ListOrders(p.Context, limit, offset)
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createOrder": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(createOrderInput),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input, ok := p.Args["input"].(map[string]interface{})
					if !ok {
						return nil, nil
					}

					customerID, okCustomer := input["customerId"].(int)
					productID, okProduct := input["productId"].(int)
					quantity, okQuantity := input["quantity"].(int)

					if !okCustomer || !okProduct || !okQuantity ||
						customerID <= 0 || productID <= 0 || quantity <= 0 {
						return nil, nil
					}

					request := model.CreateOrderRequest{
						CustomerID: int64(customerID),
						ProductID:  int64(productID),
						Quantity:   quantity,
					}

					return svc.CreateOrder(p.Context, request)
				},
			},
			"updateOrderStatus": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"status": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, okID := p.Args["id"].(int)
					status, okStatus := p.Args["status"].(string)

					if !okID || !okStatus || id <= 0 || status == "" {
						return nil, nil
					}

					return svc.UpdateOrderStatus(
						p.Context,
						int64(id),
						status,
					)
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

func Handler(schema graphql.Schema) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Query         string                 `json:"query"`
			Variables     map[string]interface{} `json:"variables"`
			OperationName string                 `json:"operationName"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []gin.H{
					{"message": "Invalid GraphQL JSON request"},
				},
			})
			return
		}

		if request.Query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []gin.H{
					{"message": "GraphQL query is required"},
				},
			})
			return
		}

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  request.Query,
			VariableValues: request.Variables,
			OperationName:  request.OperationName,
			Context:        c.Request.Context(),
		})

		c.JSON(http.StatusOK, result)
	}
}
