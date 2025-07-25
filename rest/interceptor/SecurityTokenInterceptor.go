package interceptor

import (
	"fmt"
	"net/http"

	"github.com/fantasy-versus/utils/log"
	"github.com/fantasy-versus/utils/rest/jwt"
)

// JwtInterceptor Executes all operations necessary to validate that the token received contains the credentials required for the request.
// Receives as parameter RoleValidator from the applica
func (config *ServiceConfig) JwtInterceptor(roleValidator RoleValidator, dbTokenValidator jwt.DBTokenValidator) MiddlewareInterceptor {

	return func(w http.ResponseWriter, r *http.Request, chain http.HandlerFunc) {

		if r == nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Errorln(("***ERROR***: Missing Request"))
			return
		}
		if r.Header.Get(AUTHORIZATION) == "" {
			log.Errorln(("***ERROR***: AUTHORIZATION header is missing"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := StripBearerPrefixFromTokenString(r.Header.Get(AUTHORIZATION))
		if err != nil {
			log.Errorf("Incorrect Authorization Bearer: {%s}. %+v", r.Header.Get(AUTHORIZATION), err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		c, userID, err := jwt.VerifyTokenWityIP(dbTokenValidator, &token, "")

		if err != nil {
			log.Errorf("***ERROR*** There is an error verifying the token %s. %+v", token, err)
			returnErrorToCustomer(w, NewErrorDTO("TOKEN_EXCEPTION", err.Error()))
			return
		}

		if roleValidator(r, c) {
			log.Traceln(r.UserAgent())
			r.Header.Add(HEADER_JTI, c.ID)
			r.Header.Add(HEADER_USER_NAME, c.Username)
			r.Header.Add(HEADER_CUSTOMER_ID, fmt.Sprintf("%s", c.CustomerID))
			if userID != nil {
				r.Header.Add(HEADER_USER_ID, userID.String())
			}
			r.Header.Add(HEADER_SERVICE_ID, fmt.Sprintf("%d", config.ServiceID))

			log.Infof("Request coming from user %s and customer %s", c.Username, c.CustomerID)
			chain(w, r)
		} else {
			log.Errorf("User %s is trying to access url %s with method %s with insuficient scope", c.Username, r.RequestURI, r.Method)
			returnErrorToCustomer(w, NewErrorDTO(INSUFFICIENT_SCOPE_ERROR, "Insufficient Scope"))
			return
		}

	}
}
