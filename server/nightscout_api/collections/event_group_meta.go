package collections

import (
	"context"
	"database/sql"
	"log/slog"

	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/server/nightscout_api/nsquery"
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils/httputils"
)

func EventGroupGet(ctx context.Context, siteContext *httputils.SiteContext, identifier string) (*site.SearchEventGroupsRow, *nsutils.HTTPError) {
	dbModel, err := siteContext.SiteDB.Queries.GetEventGroup(ctx, site.GetEventGroupParams{
		ID: identifier,
	})
	if err == sql.ErrNoRows || dbModel == nil {
		return nil, nsutils.ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get device status", "error", err)
		return nil, nsutils.ErrInternal
	}

	seg := site.GetEventGroupToSearchEventGroup(*dbModel)

	return &seg, nil
}

func EventGroupSearch(ctx context.Context, siteContext *httputils.SiteContext, query nsquery.Query) ([]*site.SearchEventGroupsRow, *nsutils.HTTPError) {
	if query.TypeIn == nil {
		// TODO list all supportted types so we can limit to those that NS Treatment expects
	}

	eventGroups, rerr := siteContext.SiteDB.Queries.SearchEventGroups(ctx, nsquery.EventGroupSearch(query))
	if rerr != nil {
		slog.Error("failed to query event groups from database", "error", rerr)
		return nil, nsutils.ErrInternal
	}

	return eventGroups, nil

}

func EventGroupDelete(ctx context.Context, siteContext *httputils.SiteContext, identifier string, permanent bool) *nsutils.HTTPError {
	if permanent {
		res, err := siteContext.SiteDB.Queries.HardDeleteEventGroup(ctx, site.HardDeleteEventGroupParams{ID: identifier})
		if err != nil {
			slog.Error("failed to permanently delete treatment event group by ID", "error", err)
			return nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for treatment event group delete", "error", err)
			return nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return nsutils.ErrNotFound
		}
		slog.Info("Deleted treatment event group", "eid", identifier, "rowsAffected", rowsAffected)
		return nil

	}

	res, err := siteContext.SiteDB.Queries.SoftDeleteEventGroup(ctx, site.SoftDeleteEventGroupParams{ID: identifier, UpdatedByIdentityName: &siteContext.IdentityModel.Name})
	if err != nil {
		slog.Error("failed to soft delete treatment event group by ID", "error", err)
		return nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for treatment event group delete", "error", err)
		return nsutils.ErrInternal
	}
	if rowsAffected == 0 {
		return nsutils.ErrNotFound
	}
	slog.Info("Soft Deleted treatment event group", "eid", identifier, "rowsAffected", rowsAffected)

	return nil
}

func EventGroupDeleteType(ctx context.Context, siteContext *httputils.SiteContext, typeString string, permanent bool) (int, *nsutils.HTTPError) {

	if permanent {
		res, err := siteContext.SiteDB.Queries.HardDeleteEventGroupsByType(ctx, site.HardDeleteEventGroupsByTypeParams{GroupType: typeString})
		if err != nil {
			slog.Error("failed to permanently delete treatment event group by type", "error", err)
			return 0, nsutils.ErrInternal
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			slog.Error("failed to get rows affected for treatment event group delete", "error", err)
			return 0, nsutils.ErrInternal
		}
		if rowsAffected == 0 {
			return 0, nsutils.ErrNotFound
		}
		slog.Info("Deleted treatment event group", "typeString", typeString, "rowsAffected", rowsAffected)

		return int(rowsAffected), nil

	}

	res, err := siteContext.SiteDB.Queries.SoftDeleteEventGroupsByType(ctx, site.SoftDeleteEventGroupsByTypeParams{GroupType: typeString, UpdatedByIdentityName: &siteContext.IdentityModel.Name})
	if err != nil {
		slog.Error("failed to soft delete treatment event group by type", "error", err)
		return 0, nsutils.ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for treatment event group delete", "error", err)
		return 0, nsutils.ErrInternal
	}
	if rowsAffected == 0 {
		return 0, nsutils.ErrNotFound
	}
	slog.Info("Soft Deleted treatment event group", "typeString", typeString, "rowsAffected", rowsAffected)

	return int(rowsAffected), nil
}
