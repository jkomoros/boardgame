package api

import (
	"github.com/alternaDev/go-firebase-verify"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/users"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const cookieName = "c"
const cookieLength = 64

const randomStringChars = "abcdefghijklmnopqrstuvwxyz0123456789"

//randomString returns a random string of the given length.
func randomString(length int) string {
	var result = ""

	for len(result) < length {
		result += string(randomStringChars[rand.Intn(len(randomStringChars))])
	}

	return result
}

//authCookieHandler gets the JWT and the uid and the cookie. If the given uid
//is  already tied to the given cookie, it does nothing and returns success.
//If the cookie is tied to a different uid, it barfs. If there is no UID, but
//there is a cookie, it removes that row in the DB and Set-Cookie's to remove
//the cookie. If there is no cookie, it validates the JWT, and then creates a
//new cookie tyied to that uid (creating that user record if necessary), and
//Set-Cookie's it back.
func (s *Server) authCookieHandler(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		panic("This can only be called as a post.")
	}

	uid := c.PostForm("uid")
	token := c.PostForm("token")

	cookie, _ := c.Cookie(cookieName)

	log.Println("Auth Cookie Handler called", uid, token, cookie, "*")

	//If the user is already associated with that cookie it's a success, nothing more to do.

	if cookie != "" {
		userRecord := s.storage.GetUserByCookie(cookie)

		if userRecord != nil {
			if userRecord.Id == uid {
				c.JSON(http.StatusOK, gin.H{
					"Status":  "Success",
					"Message": "Cookie and uid already matched.",
				})
				return
			}
		}
	}

	if uid == "" && cookie != "" {
		//We must have an old cookie set. Clear it out.
		if err := s.storage.ConnectCookieToUser(cookie, nil); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": "Failure",
				"Error":  err.Error(),
			})
			return
		}

		//Delete the cookie on the client.

		//TODO: might need to set the domain in production.
		c.SetCookie(cookieName, "", int(time.Now().Add(time.Hour*10000*-1).Unix()), "", "", false, false)

		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Removed cookie for signed-out uid",
		})

		return

	}

	if cookie == "" && uid != "" {

		verifiedUid, err := firebase.VerifyIDToken(token, "boardgame-159316")

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": "Failure",
				"Error":  "Failed to verify jwt token: " + err.Error(),
			})
			return
		}

		if verifiedUid != uid {
			c.JSON(http.StatusOK, gin.H{
				"Status": "Failure",
				"Error":  "The decoded jwt token doesn not match with the provided uid.",
			})
			return
		}

		user := s.storage.GetUserById(uid)

		//If we've never seen this Uid before, store it.
		if user == nil {

			user = &users.StorageRecord{
				Id: uid,
			}
			s.storage.UpdateUser(user)

		}

		cookie = randomString(cookieLength)

		if err := s.storage.ConnectCookieToUser(cookie, user); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": "Failure",
				"Error":  "Couldn't connect cookie to user: " + err.Error(),
			})
			return
		}

		//TODO: might need to set the domain in production
		c.SetCookie(cookieName, cookie, int(time.Now().Add(time.Hour*100).Unix()), "", "", false, false)

		c.JSON(http.StatusOK, gin.H{
			"Status":  "Success",
			"Message": "Created new cookie to point to uid",
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"Status": "Failure",
		"Error":  "Not Yet Implemented",
	})
}
