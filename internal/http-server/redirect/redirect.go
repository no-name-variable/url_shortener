package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url_shortner/internal/lib/api/response"
	"url_shortner/internal/models"
	"url_shortner/internal/storage"
	"url_shortner/internal/storage/lib/logger/sl"
)

type URLGetter interface {
	GetURLByAlias(Alias string) (models.URL, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op))
		slog.String("request_id", middleware.GetReqID(r.Context()))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		resUrl, err := urlGetter.GetURLByAlias(alias)
		if err != nil {
			var msgErr = "alias not found"
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info(msgErr, sl.Err(err))
				render.JSON(w, r, response.Error(msgErr))
				return
			}
			log.Error("unknown err", sl.Err(err))
			render.JSON(w, r, response.Error("unknown error"))
			return
		}
		redirectUrl := resUrl.Url
		log.Info("got url", slog.String("url", redirectUrl))
		http.Redirect(w, r, redirectUrl, http.StatusFound)

		return
	}
}
