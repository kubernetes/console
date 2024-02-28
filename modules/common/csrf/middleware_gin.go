package csrf

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/xsrftoken"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/errors"
)

var (
	ginMiddlewares = &GinMiddlewares{}
)

func Gin() *GinMiddlewares {
	return ginMiddlewares
}

type GinMiddlewares struct{}

func (in *GinMiddlewares) CSRF(options ...GinCSRFOption) gin.HandlerFunc {
	fmt.Println("registering new middleware")
	middleware := &GinCSRFMiddleware{
		actionGetter: defaultGinCSRFActionGetter,
		runCondition: defaultGinCSRFRunCondition,
	}

	for _, opt := range options {
		opt(middleware)
	}

	return middleware.build()
}

func (in *GinMiddlewares) WithCSRFActionGetter(getter GinCSRFActionGetter) GinCSRFOption {
	return func(middleware *GinCSRFMiddleware) {
		middleware.actionGetter = getter
	}
}

func (in *GinMiddlewares) WithCSRFRunCondition(condition GinCSRFRunCondition) GinCSRFOption {
	return func(middleware *GinCSRFMiddleware) {
		middleware.runCondition = condition
	}
}

type GinCSRFOption func(middleware *GinCSRFMiddleware)
type GinCSRFActionGetter func(selectedRoutePath string) (action *string)
type GinCSRFRunCondition func(request *http.Request) bool

var (
	defaultGinCSRFActionGetter = func(selectedRoutePath string) *string {
		return &selectedRoutePath
	}

	defaultGinCSRFRunCondition = func(request *http.Request) bool {
		return request.Method == http.MethodPost
	}
)

type GinCSRFMiddleware struct {
	actionGetter GinCSRFActionGetter
	runCondition GinCSRFRunCondition
}

func (in *GinCSRFMiddleware) build() gin.HandlerFunc {
	return func(c *gin.Context) {
		actionID := in.actionGetter(c.Request.URL.Path)

		if !in.runCondition(c.Request) {
			c.Next()
			return
		}

		klog.V(4).InfoS("[GinCSRFMiddleware] Got request", "path", c.Request.URL.Path, "actionID", actionID)
		if actionID == nil || !xsrftoken.Valid(c.Request.Header.Get(csrfTokenHeader), Key(), "none", *actionID) {
			klog.Errorf("CSRF validation failed, actionID: %s", *actionID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.NewCSRFValidationFailed())
			return
		}

		c.Next()
	}
}
