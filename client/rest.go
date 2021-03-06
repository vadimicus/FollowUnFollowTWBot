package client

import (
	"github.com/vadimicus/FollowUnFollowTWBot/store"
	"github.com/gin-gonic/gin"
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

	r.GET("/ping", restClient.Ping())

	return restClient, nil
}

func (restClient *RestClient) Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		all,_ := restClient.userStore.GetAllUsers()
		to_follow,_ := restClient.userStore.GetUsersToFollow()
		to_unfollow, _ := restClient.userStore.GetUsersToUnFollow()
		done, _ := restClient.userStore.GetUsersByStatus(2)
		followError, _ := restClient.userStore.GetUsersByStatus(66)
		unFollowError, _ := restClient.userStore.GetUsersByStatus(666)


		resp := map[string]interface{}{}

		//resp := map[string]interface{}{
		//	"stockexchanges": map[string][]string{
		//		"poloniex": []string{"usd_btc", "eth_btc", "eth_usd", "btc_usd"},
		//		"gdax":     []string{"eur_btc", "usd_btc", "eth_btc", "eth_usd", "eth_eur", "btc_usd"},
		//	},
		//	"servertime": time.Now().UTC().Unix(),
		//	"api":        "1.2",
		//	//"version":    restClient.MultyVersion,
		//	//"donate":     restClient.donationAddresses,
		//	"multisigfactory": map[string]string{
		//		"ethtestnet": "0x04f68589f53cfdf408025cd7cea8a40dbf488e49",
		//		"ethmainnet": "0xc2cbdd9b58502cff1db5f9cce48ac17a9a550185",
		//	},
		//	//"erc20tokenlist": restClient.ERC20TokenList,
		//}
		resp["users"] = map[string]int{
			"all": len(all),
			"to_follow": len(to_follow),
			"to_unfollow": len(to_unfollow),
			"done": len(done),
			"error_to_follow": len(followError),
			"error_to_unfollow": len(unFollowError),
		}



		c.JSON(http.StatusOK, resp)
	}
}