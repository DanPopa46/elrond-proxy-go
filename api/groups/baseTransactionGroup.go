package groups

import (
	"fmt"
	"net/http"

	"github.com/ElrondNetwork/elrond-proxy-go/api/errors"
	"github.com/ElrondNetwork/elrond-proxy-go/api/shared"
	"github.com/ElrondNetwork/elrond-proxy-go/data"
	"github.com/gin-gonic/gin"
)

type transactionGroup struct {
	facade TransactionFacadeHandler
	*baseGroup
}

// NewNodeGroup returns a new instance of nodeGroup
func NewTransactionGroup(facadeHandler data.FacadeHandler) (*transactionGroup, error) {
	facade, ok := facadeHandler.(TransactionFacadeHandler)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	tg := &transactionGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	baseRoutesHandlers := map[string]*data.EndpointHandlerData{
		"/send":            {Handler: tg.SendTransaction, Method: http.MethodPost},
		"/simulate":        {Handler: tg.SimulateTransaction, Method: http.MethodPost},
		"/send-multiple":   {Handler: tg.SendMultipleTransactions, Method: http.MethodPost},
		"/send-user-funds": {Handler: tg.SendUserFunds, Method: http.MethodPost},
		"/cost":            {Handler: tg.RequestTransactionCost, Method: http.MethodPost},
		"/:txhash/status":  {Handler: tg.GetTransactionStatus, Method: http.MethodGet},
		"/:txhash":         {Handler: tg.GetTransaction, Method: http.MethodGet},
	}
	tg.baseGroup.endpoints = baseRoutesHandlers

	return tg, nil
}

// SendTransaction will receive a transaction from the client and propagate it for processing
func (tg *transactionGroup) SendTransaction(c *gin.Context) {
	var tx = data.Transaction{}
	err := c.ShouldBindJSON(&tx)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusBadRequest,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
			data.ReturnCodeRequestError,
		)
		return
	}

	statusCode, txHash, err := tg.facade.SendTransaction(&tx)
	if err != nil {
		shared.RespondWith(c, statusCode, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"txHash": txHash}, "", data.ReturnCodeSuccess)
}

// SendUserFunds will receive an address from the client and propagate a transaction for sending some ERD to that address
func (tg *transactionGroup) SendUserFunds(c *gin.Context) {
	if !tg.facade.IsFaucetEnabled() {
		shared.RespondWith(
			c,
			http.StatusBadRequest,
			nil,
			errors.ErrFaucetNotEnabled.Error(),
			data.ReturnCodeRequestError,
		)
		return
	}

	var gtx = data.FundsRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusBadRequest,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
			data.ReturnCodeRequestError,
		)
		return
	}

	err = tg.facade.SendUserFunds(gtx.Receiver, gtx.Value)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error()),
			data.ReturnCodeRequestError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"message": "ok"}, "", data.ReturnCodeSuccess)
}

// SendMultipleTransactions will send multiple transactions at once
func (tg *transactionGroup) SendMultipleTransactions(c *gin.Context) {
	var txs []*data.Transaction
	err := c.ShouldBindJSON(&txs)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusBadRequest,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
			data.ReturnCodeRequestError,
		)
		return
	}

	response, err := tg.facade.SendMultipleTransactions(txs)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error()),
			data.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(
		c,
		http.StatusOK,
		gin.H{
			"numOfSentTxs": response.NumOfTxs,
			"txsHashes":    response.TxsHashes,
		},
		"",
		data.ReturnCodeSuccess,
	)
}

// SimulateTransaction will receive a transaction from the client and will send it for simulation purpose
func (tg *transactionGroup) SimulateTransaction(c *gin.Context) {
	var tx = data.Transaction{}
	err := c.ShouldBindJSON(&tx)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusBadRequest,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
			data.ReturnCodeRequestError,
		)
		return
	}

	simulationResponse, err := tg.facade.SimulateTransaction(&tx)
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	c.JSON(
		http.StatusOK,
		simulationResponse,
	)
}

// RequestTransactionCost will return an estimation of how many gas unit a transaction will cost
func (tg *transactionGroup) RequestTransactionCost(c *gin.Context) {
	var tx = data.Transaction{}
	err := c.ShouldBindJSON(&tx)
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusBadRequest,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error()),
			data.ReturnCodeInternalError,
		)
		return
	}

	cost, err := tg.facade.TransactionCostRequest(&tx)
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"txGasUnits": cost}, "", data.ReturnCodeSuccess)
}

// GetTransactionStatus will return the transaction's status
func (tg *transactionGroup) GetTransactionStatus(c *gin.Context) {
	txHash := c.Param("txhash")
	sender := c.Request.URL.Query().Get("sender")
	txStatus, err := tg.facade.GetTransactionStatus(txHash, sender)
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"status": txStatus}, "", data.ReturnCodeSuccess)
}

// GetTransaction should return a transaction from observer
func (tg *transactionGroup) GetTransaction(c *gin.Context) {
	txHash := c.Param("txhash")
	if txHash == "" {
		shared.RespondWith(c, http.StatusBadRequest, nil, errors.ErrTransactionHashMissing.Error(), data.ReturnCodeRequestError)
		return
	}

	sndAddr := c.Request.URL.Query().Get("sender")
	if sndAddr != "" {
		getTransactionByHashAndSenderAddress(c, tg.facade, txHash, sndAddr)
		return
	}

	tx, err := tg.facade.GetTransaction(txHash)
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"transaction": tx}, "", data.ReturnCodeSuccess)
}

func getTransactionByHashAndSenderAddress(c *gin.Context, ef TransactionFacadeHandler, txHash string, sndAddr string) {
	tx, statusCode, err := ef.GetTransactionByHashAndSenderAddress(txHash, sndAddr)
	if err != nil {
		internalCode := data.ReturnCodeInternalError
		if statusCode == http.StatusBadRequest {
			internalCode = data.ReturnCodeRequestError
		}
		shared.RespondWith(c, statusCode, nil, err.Error(), internalCode)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"transaction": tx}, "", data.ReturnCodeSuccess)
}
