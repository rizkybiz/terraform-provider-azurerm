package authorizations

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListResponse struct {
	HttpResponse *http.Response
	Model        *[]ExpressRouteAuthorization

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (ListResponse, error)
}

type ListCompleteResult struct {
	Items []ExpressRouteAuthorization
}

func (r ListResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r ListResponse) LoadMore(ctx context.Context) (resp ListResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

type ExpressRouteAuthorizationPredicate struct {
	// TODO: implement me
}

func (p ExpressRouteAuthorizationPredicate) Matches(input ExpressRouteAuthorization) bool {
	// TODO: implement me
	// if p.Name != nil && input.Name != *p.Name {
	// 	return false
	// }

	return true
}

// List ...
func (c AuthorizationsClient) List(ctx context.Context, id PrivateCloudId) (resp ListResponse, err error) {
	req, err := c.preparerForList(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "authorizations.AuthorizationsClient", "List", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "authorizations.AuthorizationsClient", "List", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForList(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "authorizations.AuthorizationsClient", "List", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// ListCompleteMatchingPredicate retrieves all of the results into a single object
func (c AuthorizationsClient) ListComplete(ctx context.Context, id PrivateCloudId) (ListCompleteResult, error) {
	return c.ListCompleteMatchingPredicate(ctx, id, ExpressRouteAuthorizationPredicate{})
}

// ListCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c AuthorizationsClient) ListCompleteMatchingPredicate(ctx context.Context, id PrivateCloudId, predicate ExpressRouteAuthorizationPredicate) (resp ListCompleteResult, err error) {
	items := make([]ExpressRouteAuthorization, 0)

	page, err := c.List(ctx, id)
	if err != nil {
		err = fmt.Errorf("loading the initial page: %+v", err)
		return
	}
	if page.Model != nil {
		for _, v := range *page.Model {
			if predicate.Matches(v) {
				items = append(items, v)
			}
		}
	}

	for page.HasMore() {
		page, err = page.LoadMore(ctx)
		if err != nil {
			err = fmt.Errorf("loading the next page: %+v", err)
			return
		}

		if page.Model != nil {
			for _, v := range *page.Model {
				if predicate.Matches(v) {
					items = append(items, v)
				}
			}
		}
	}

	out := ListCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForList prepares the List request.
func (c AuthorizationsClient) preparerForList(ctx context.Context, id PrivateCloudId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/authorizations", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForListWithNextLink prepares the List request with the given nextLink token.
func (c AuthorizationsClient) preparerForListWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
	uri, err := url.Parse(nextLink)
	if err != nil {
		return nil, fmt.Errorf("parsing nextLink %q: %+v", nextLink, err)
	}
	queryParameters := map[string]interface{}{}
	for k, v := range uri.Query() {
		if len(v) == 0 {
			continue
		}
		val := v[0]
		val = autorest.Encode("query", val)
		queryParameters[k] = val
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(uri.Path),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForList handles the response to the List request. The method always
// closes the http.Response Body.
func (c AuthorizationsClient) responderForList(resp *http.Response) (result ListResponse, err error) {
	type page struct {
		Values   []ExpressRouteAuthorization `json:"value"`
		NextLink *string                     `json:"nextLink"`
	}
	var respObj page
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&respObj),
		autorest.ByClosing())
	result.HttpResponse = resp
	result.Model = &respObj.Values
	result.nextLink = respObj.NextLink
	if respObj.NextLink != nil {
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result ListResponse, err error) {
			req, err := c.preparerForListWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "authorizations.AuthorizationsClient", "List", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "authorizations.AuthorizationsClient", "List", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForList(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "authorizations.AuthorizationsClient", "List", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}
