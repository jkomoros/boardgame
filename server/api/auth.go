package api

import (
	"github.com/alternaDev/go-firebase-verify"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/server/api/users"
	"math/rand"
	"net/http"
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

func (s *Server) unsetCookie(r *Renderer, cookie string, message string) {
	//We must have an old cookie set. Clear it out.
	if err := s.storage.ConnectCookieToUser(cookie, nil); err != nil {

		r.Error(err.Error())
		return
	}

	//Delete the cookie on the client.
	r.SetAuthCookie("")

	r.Success(gin.H{
		"Message": message,
	})
	return
}

//authCookieHandler gets the JWT and the uid and the cookie. If the given uid
//is  already tied to the given cookie, it does nothing and returns success.
//If the cookie is tied to a different uid, it barfs. If there is no UID, but
//there is a cookie, it removes that row in the DB and Set-Cookie's to remove
//the cookie. If there is no cookie, it validates the JWT, and then creates a
//new cookie tyied to that uid (creating that user record if necessary), and
//Set-Cookie's it back.
func (s *Server) authCookieHandler(c *gin.Context) {

	r := NewRenderer(c)

	if c.Request.Method != http.MethodPost {
		r.Error("This method only supports post.")
		return
	}

	uid := c.PostForm("uid")
	token := c.PostForm("token")

	cookie, _ := c.Cookie(cookieName)

	//If the user is already associated with that cookie it's a success, nothing more to do.

	if cookie != "" && uid != "" {

		userRecord := s.storage.GetUserByCookie(cookie)

		if userRecord == nil {
			//The cookie must be invalid
			s.unsetCookie(r, cookie, "Cookie pointed to an user that did not exist. Unsetting.")
			return
		} else {
			if userRecord.Id == uid {

				r.Success(gin.H{
					"Message": "Cookie and uid already matched.",
				})
				return
			} else {
				s.unsetCookie(r, cookie, "Cookie pointed to the wrong uid. Unsetting")
				return
			}
		}
	}

	if uid == "" && cookie != "" {
		s.unsetCookie(r, cookie, "Removed cookie for signed-out uid")
		return
	}

	if cookie == "" && uid != "" {

		verifiedUid, err := firebase.VerifyIDToken(token, s.config.FirebaseProjectId)

		if err != nil {
			r.Error("Failed to verify jwt token: " + err.Error())
			return
		}

		if verifiedUid != uid {

			r.Error("The decoded jwt token doesn not match with the provided uid.")
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

			r.Error("Couldn't connect cookie to user: " + err.Error())
			return
		}

		r.SetAuthCookie(cookie)

		r.Success(gin.H{
			"Message": "Created new cookie to point to uid",
		})

		return

	}

	r.Error("Unexpectedly reached end of function")

}
