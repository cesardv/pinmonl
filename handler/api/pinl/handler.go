package pinl

import (
	"bytes"
	"context"
	"net/http"

	"github.com/pinmonl/pinmonl/handler/api/apiutils"
	"github.com/pinmonl/pinmonl/handler/api/image"
	"github.com/pinmonl/pinmonl/handler/api/request"
	"github.com/pinmonl/pinmonl/handler/api/response"
	"github.com/pinmonl/pinmonl/model"
	"github.com/pinmonl/pinmonl/pkg/scrape"
	"github.com/pinmonl/pinmonl/queue"
	"github.com/pinmonl/pinmonl/store"
)

// HandleList returns pinls.
func HandleList(pinls store.PinlStore, tags store.TagStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := request.UserFrom(ctx)
		ms, err := pinls.List(ctx, &store.PinlOpts{UserID: u.ID})
		if err != nil {
			response.InternalError(w, err)
			return
		}

		mps := model.MustBeMorphables(ms)
		ts, err := tags.List(ctx, &store.TagOpts{Targets: mps})
		if err != nil {
			response.InternalError(w, err)
			return
		}

		resp := make([]interface{}, len(ms))
		for i, m := range ms {
			mt := (model.TagList)(ts).FindMorphable(m)
			resp[i] = Resp(m, mt)
		}
		response.JSON(w, resp)
	}
}

// HandleFind returns pinl and its relations.
func HandleFind(tags store.TagStore, pkgs store.PkgStore, stats store.StatStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		m, has := request.PinlFrom(ctx)
		if !has {
			response.NotFound(w, nil)
			return
		}

		ts, err := tags.List(ctx, &store.TagOpts{Target: m})
		if err != nil {
			response.InternalError(w, err)
			return
		}

		ms, err := getPkgsFromURL(ctx, pkgs, m.URL)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		mss, err := getStats(ctx, stats, ms)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		response.JSON(w, DetailResp(m, ts, ms, mss))
	}
}

// HandleCreate validates and create pinl from user input.
func HandleCreate(
	pinls store.PinlStore,
	tags store.TagStore,
	taggables store.TaggableStore,
	qm *queue.Manager,
	images store.ImageStore,
	pkgs store.PkgStore,
	stats store.StatStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		in, err := ReadInput(r.Body)
		if err != nil {
			response.BadRequest(w, err)
			return
		}

		if err = in.Validate(); err != nil {
			response.BadRequest(w, err)
			return
		}

		ctx := r.Context()
		u, _ := request.UserFrom(ctx)
		m := model.Pinl{UserID: u.ID}
		err = in.Fill(&m)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		err = fillCardIfEmpty(ctx, images, &m)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		err = pinls.Create(ctx, &m)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		ts, err := apiutils.FindOrCreateTagsByName(ctx, tags, u, in.Tags)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		err = taggables.ReAssocTags(ctx, m, ts)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		err = qm.Enqueue(ctx, &model.Job{
			Name:     model.JobPinlCreated,
			TargetID: m.ID,
		})
		if err != nil {
			response.InternalError(w, err)
			return
		}

		ms, err := getPkgsFromURL(ctx, pkgs, m.URL)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		mss, err := getStats(ctx, stats, ms)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		response.JSON(w, DetailResp(m, ts, ms, mss))
	}
}

// HandleUpdate validates and updates pinl from user input.
func HandleUpdate(
	pinls store.PinlStore,
	tags store.TagStore,
	taggables store.TaggableStore,
	qm *queue.Manager,
	images store.ImageStore,
	pkgs store.PkgStore,
	stats store.StatStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		in, err := ReadInput(r.Body)
		if err != nil {
			response.BadRequest(w, err)
			return
		}

		if err = in.Validate(); err != nil {
			response.BadRequest(w, err)
			return
		}

		ctx := r.Context()
		u, _ := request.UserFrom(ctx)
		m, _ := request.PinlFrom(ctx)
		err = in.Fill(&m)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		err = fillCardIfEmpty(ctx, images, &m)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		err = pinls.Update(ctx, &m)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		ts, err := apiutils.FindOrCreateTagsByName(ctx, tags, u, in.Tags)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		err = taggables.ReAssocTags(ctx, m, ts)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		err = qm.Enqueue(ctx, &model.Job{
			Name:     model.JobPinlUpdated,
			TargetID: m.ID,
		})
		if err != nil {
			response.InternalError(w, err)
			return
		}

		ms, err := getPkgsFromURL(ctx, pkgs, m.URL)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		mss, err := getStats(ctx, stats, ms)
		if err != nil {
			response.InternalError(w, err)
			return
		}

		response.JSON(w, DetailResp(m, ts, ms, mss))
	}
}

// HandleDelete removes pinl and its relations.
func HandleDelete(
	pinls store.PinlStore,
	taggables store.TaggableStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		m, _ := request.PinlFrom(ctx)

		err := pinls.Delete(ctx, &m)
		if err != nil {
			response.InternalError(w, nil)
			return
		}

		err = taggables.ClearTags(ctx, m)
		if err != nil {
			response.InternalError(w, nil)
			return
		}

		response.NoContent(w)
	}
}

// HandlePageInfo returns the page info of Pinl.
func HandlePageInfo(pinls store.PinlStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := request.UserFrom(ctx)
		count, err := pinls.Count(ctx, &store.PinlOpts{UserID: u.ID})
		if err != nil {
			response.InternalError(w, err)
			return
		}

		pi := response.PageInfo{
			Count: count,
		}
		response.JSON(w, pi)
	}
}

func fillCardIfEmpty(ctx context.Context, images store.ImageStore, m *model.Pinl) error {
	if m.Title != "" {
		return nil
	}

	resp, err := scrape.Get(m.URL)
	if err != nil {
		return err
	}
	card, err := resp.Card()
	if err != nil {
		return err
	}
	ci, err := card.Image()
	if err != nil {
		return err
	}

	m2 := *m
	m2.Title = card.Title()
	m2.Description = card.Description()
	if ci != nil {
		img, err := image.UploadFromReader(ctx, images, bytes.NewBuffer(ci))
		if err != nil {
			return err
		}
		m2.ImageID = img.ID
	}
	*m = m2
	return nil
}

func getPkgsFromURL(ctx context.Context, pkgs store.PkgStore, url string) ([]model.Pkg, error) {
	ms, err := pkgs.List(ctx, &store.PkgOpts{MonlURL: url})
	if err != nil {
		return nil, err
	}
	return ms, nil
}

func getStats(ctx context.Context, stats store.StatStore, pkgs []model.Pkg) ([]model.Stat, error) {
	if len(pkgs) == 0 {
		return []model.Stat{}, nil
	}
	pids := (model.PkgList)(pkgs).Keys()
	ss, err := stats.List(ctx, &store.StatOpts{PkgIDs: pids, WithLatest: true})
	if err != nil {
		return nil, err
	}
	return ss, nil
}
