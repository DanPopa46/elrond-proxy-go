package shared

import (
	"net/http"
	"strconv"

	"github.com/ElrondNetwork/elrond-proxy-go/api/errors"
	"github.com/ElrondNetwork/elrond-proxy-go/data"
	"github.com/gin-gonic/gin"
)

// RespondWith will respond with the generic API response
func RespondWith(c *gin.Context, status int, dataField interface{}, error string, code data.ReturnCode) {
	c.JSON(
		status,
		data.GenericAPIResponse{
			Data:  dataField,
			Error: error,
			Code:  code,
		},
	)
}

// FetchNonceFromRequest will try to fetch the nonce from the request
func FetchNonceFromRequest(c *gin.Context) (uint64, error) {
	nonceStr := c.Param("nonce")
	if nonceStr == "" {
		return 0, errors.ErrInvalidBlockNonceParam
	}

	return strconv.ParseUint(nonceStr, 10, 64)
}

// FetchShardIDFromRequest will try to fetch the shard ID from the request
func FetchShardIDFromRequest(c *gin.Context) (uint32, error) {
	shardStr := c.Param("shard")
	if shardStr == "" {
		return 0, errors.ErrInvalidShardIDParam
	}

	shardID, err := strconv.ParseUint(shardStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(shardID), nil
}

// RespondWithBadRequest creates a generic response for bad request
func RespondWithBadRequest(c *gin.Context, errorMessage string) {
	RespondWith(c, http.StatusBadRequest, nil, errorMessage, data.ReturnCodeRequestError)
}
