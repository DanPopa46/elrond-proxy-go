package api

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/ElrondNetwork/elrond-proxy-go/api/address"
	"github.com/ElrondNetwork/elrond-proxy-go/api/block"
	"github.com/ElrondNetwork/elrond-proxy-go/api/network"
	"github.com/ElrondNetwork/elrond-proxy-go/api/node"
	"github.com/ElrondNetwork/elrond-proxy-go/api/transaction"
	valStats "github.com/ElrondNetwork/elrond-proxy-go/api/validator"
	"github.com/ElrondNetwork/elrond-proxy-go/api/vmValues"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// CreateServer creates a HTTP server
func CreateServer(elrondProxyFacade ElrondProxyHandler, port int) (*http.Server, error) {
	ws := gin.Default()
	ws.Use(cors.Default())

	err := registerValidators()
	if err != nil {
		return nil, err
	}

	registerRoutes(ws, elrondProxyFacade)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: ws,
	}

	return httpServer, nil
}

func registerRoutes(ws *gin.Engine, elrondProxyFacade ElrondProxyHandler) {
	addressRoutes := ws.Group("/address")
	addressRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	address.Routes(addressRoutes)

	txRoutes := ws.Group("/transaction")
	txRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	transaction.Routes(txRoutes)

	getValuesRoutes := ws.Group("/vm-values")
	getValuesRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	vmValues.Routes(getValuesRoutes)

	networkRoutes := ws.Group("/network")
	networkRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	network.Routes(networkRoutes)

	nodeRoutes := ws.Group("/node")
	nodeRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	node.Routes(nodeRoutes)

	validatorRoutes := ws.Group("/validator")
	validatorRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	valStats.Routes(validatorRoutes)

	blockRoutes := ws.Group("/block")
	blockRoutes.Use(WithElrondProxyFacade(elrondProxyFacade))
	block.Routes(blockRoutes)
}

func registerValidators() error {
	validators := []validatorInput{
		{Name: "skValidator", Validator: skValidator},
	}
	for _, validatorFunc := range validators {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			err := v.RegisterValidation(validatorFunc.Name, validatorFunc.Validator)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// skValidator validates a secret key from user input for correctness
func skValidator(
	_ *validator.Validate,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Type,
	_ reflect.Kind,
	_ string,
) bool {
	return true
}

// WithElrondProxyFacade middleware will set up an ElrondFacade object in the gin context
func WithElrondProxyFacade(elrondProxyFacade ElrondProxyHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("elrondProxyFacade", elrondProxyFacade)
		c.Next()
	}
}
