package api

import (
	"github.com/fazilb93/neutron/backend"
	"gopkg.in/macaron.v1"
)

type AvailableDomainsResp struct {
	Resp
	Domains []string
}

func (api *Api) GetAvailableDomains(ctx *macaron.Context) (err error) {
	domains, err := api.backend.ListDomains()
	if err != nil {
		return
	}

	domainNames := make([]string, len(domains))
	for i, d := range domains {
		domainNames[i] = d.DomainName
	}

	ctx.JSON(200, &AvailableDomainsResp{
		Resp:    Resp{Ok},
		Domains: domainNames,
	})
	return
}

func populateDomain(domain *backend.Domain) {
	domain.State = 1
	domain.VerifyState = 2
}

type DomainResp struct {
	Resp
	Domain *backend.Domain
}

func (api *Api) GetDomain(ctx *macaron.Context) (err error) {
	domainId := ctx.Params("id")

	domain, err := api.backend.GetDomain(domainId)
	if err != nil {
		return
	}

	populateDomain(domain)

	ctx.JSON(200, &DomainResp{
		Resp:   Resp{Ok},
		Domain: domain,
	})
	return
}

type DomainsResp struct {
	Resp
	Domains []*backend.Domain
}

func (api *Api) GetUserDomains(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)

	domains, err := api.backend.ListDomains()
	if err != nil {
		return
	}

	for _, dom := range domains {
		populateDomain(dom)
		dom.Addresses = nil
	}

	addresses, err := api.backend.ListAddresses(userId)
	if err != nil {
		return
	}

	for _, addr := range addresses {
		for _, dom := range domains {
			if dom.ID != addr.DomainID {
				continue
			}

			dom.Addresses = append(dom.Addresses, addr)
			break
		}
	}

	ctx.JSON(200, &DomainsResp{
		Resp:    Resp{Ok},
		Domains: domains,
	})
	return
}
