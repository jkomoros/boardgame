package api

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/alternaDev/go-firebase-verify"
	"github.com/gin-gonic/gin"
	"github.com/jkomoros/boardgame/errors"
	"github.com/jkomoros/boardgame/server/api/users"
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

func (s *Server) unsetCookie(r *renderer, cookie string, message string) {
	//We must have an old cookie set. Clear it out.
	if err := s.storage.ConnectCookieToUser(cookie, nil); err != nil {

		r.Error(errors.New(err.Error()))
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

	r := s.newRenderer(c)

	if c.Request.Method != http.MethodPost {
		r.Error(errors.New("this method only supports post"))
		return
	}

	uid := c.PostForm("uid")
	token := c.PostForm("token")
	email := c.PostForm("email")
	photoURL := c.PostForm("photo")
	displayName := c.PostForm("displayname")

	cookie, _ := c.Cookie(cookieName)

	s.doAuthCookie(r, uid, token, cookie, email, photoURL, displayName)

}

func (s *Server) authSuccess(r *renderer, user *users.StorageRecord, message string) {

	if user != nil {
		//Make a copy so tha twe don't overwrite the user storage record and
		//accidentally persist this EffectiveDisplayName to disk.
		var userCopy users.StorageRecord

		userCopy = *user

		user = &userCopy

		if user.DisplayName == "" {
			user.DisplayName = user.EffectiveDisplayName()
		}
	}

	adminAllowed := s.calcAdminAllowed(user)

	r.Success(gin.H{
		"User":         user,
		"AdminAllowed": adminAllowed,
		"Message":      message,
	})

}

func (s *Server) doAuthCookie(r *renderer, uid, token, cookie, email, photoURL, displayName string) {
	//If the user is already associated with that cookie it's a success, nothing more to do.

	if cookie != "" && uid != "" {

		userRecord := s.storage.GetUserByCookie(cookie)

		if userRecord == nil {
			//The cookie must be invalid; perhaps we have reset the database.

			//Unset the cookie in the database
			s.storage.ConnectCookieToUser(cookie, nil)

			//Tell the renderer to unset the cookie
			r.SetAuthCookie("")

			//Tell the rest of this handler to pretend there is no cookie,
			//which will likely sign us in.
			cookie = ""

			//Do NOT return; fall through to the rest of handler.

		} else {
			if userRecord.ID == uid {

				if userRecord.PhotoURL == "" && photoURL != "" {
					userRecord.PhotoURL = photoURL
				}

				if userRecord.DisplayName == "" && displayName != "" {
					userRecord.DisplayName = displayName
				}

				if userRecord.Email == "" && email != "" {
					userRecord.Email = email
				}

				userRecord.LastSeen = time.Now().UnixNano()

				s.storage.UpdateUser(userRecord)

				s.authSuccess(r, userRecord, "Cookie and uid already matched.")
				return
			}
			s.unsetCookie(r, cookie, "Cookie pointed to the wrong uid. Unsetting")
			return

		}
	}

	if uid == "" && cookie != "" {
		s.unsetCookie(r, cookie, "Removed cookie for signed-out uid")
		return
	}

	if cookie == "" && uid != "" {

		if s.config.OfflineDevMode {

			s.logger.Warnln("Skipping auth checking because of OfflineDevMode. This setting should NEVER be enabled in prod.")

		} else {

			verifiedUID, err := firebase.VerifyIDToken(token, s.config.Firebase.ProjectID)

			if err != nil {
				r.Error(errors.New("Failed to verify jwt token: " + err.Error()))
				return
			}

			if verifiedUID != uid {

				r.Error(errors.New("the decoded jwt token doesn not match with the provided uid"))
				return
			}
		}

		user := s.storage.GetUserByID(uid)

		//If we've never seen this Uid before, store it.
		if user == nil {

			user = &users.StorageRecord{
				ID:          uid,
				Email:       email,
				PhotoURL:    photoURL,
				DisplayName: displayName,
				Created:     time.Now().UnixNano(),
				LastSeen:    time.Now().UnixNano(),
			}
			s.storage.UpdateUser(user)

		}

		cookie = randomString(cookieLength)

		if err := s.storage.ConnectCookieToUser(cookie, user); err != nil {

			r.Error(errors.New("Couldn't connect cookie to user: " + err.Error()))
			return
		}

		r.SetAuthCookie(cookie)

		s.authSuccess(r, user, "Created new cookie to point to uid")

		return

	}

	r.Success(gin.H{
		"Message": "Not logged in, but no info passed.",
	})

}
