package server

import "avito/internal/generated"
import "avito/internal/server/middleware"

func RegisterHandlersWithAuth(router generated.EchoRouter, si generated.ServerInterface) {
	wrapper := generated.ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/banner", wrapper.GetBanner, middleware.AdminMiddleware)
	router.POST("/banner", wrapper.PostBanner, middleware.AdminMiddleware)
	router.DELETE("/banner/:id", wrapper.DeleteBannerId, middleware.AdminMiddleware)
	router.PATCH("/banner/:id", wrapper.PatchBannerId, middleware.AdminMiddleware)
	router.GET("/user_banner", wrapper.GetUserBanner, middleware.UserMiddleware)
}
