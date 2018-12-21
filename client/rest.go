package client

import (
	"github.com/vadimicus/FollowUnFollowTWBot/store"
	"github.com/gin-gonic/gin"
	"time"
	"net/http"
)

type RestClient struct {
	//middlewareJWT *GinJWTMiddleware
	userStore		store.UserStore

}

func SetRestHandlers(
	userDB store.UserStore,
	r *gin.Engine,

) (*RestClient, error) {
	restClient := &RestClient{
		userStore:userDB,
	}

	r.GET("/test", restClient.Test())

	return restClient, nil
}

func (restClient *RestClient) Test() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := map[string]interface{}{
			"stockexchanges": map[string][]string{
				"poloniex": []string{"usd_btc", "eth_btc", "eth_usd", "btc_usd"},
				"gdax":     []string{"eur_btc", "usd_btc", "eth_btc", "eth_usd", "eth_eur", "btc_usd"},
			},
			"servertime": time.Now().UTC().Unix(),
			"api":        "1.2",
			//"version":    restClient.MultyVersion,
			//"donate":     restClient.donationAddresses,
			"multisigfactory": map[string]string{
				"ethtestnet": "0x04f68589f53cfdf408025cd7cea8a40dbf488e49",
				"ethmainnet": "0xc2cbdd9b58502cff1db5f9cce48ac17a9a550185",
			},
			//"erc20tokenlist": restClient.ERC20TokenList,
		}
		resp["android"] = map[string]int{
			"soft": 1,
			"hard": 1,
		}
		resp["ios"] = map[string]int{
			"soft": 2,
			"hard": 2,
		}


		c.JSON(http.StatusOK, resp)
	}
}