// Copyright © 2018 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diva

import (
	"github.com/clearlinux/diva/download"
	"github.com/clearlinux/diva/internal/config"
	"github.com/clearlinux/diva/internal/helpers"
	"github.com/clearlinux/diva/pkginfo"
)

// DownloadRepo downloads the RPM repo to the local cache location
func DownloadRepo(conf *config.Config, u *config.UInfo, repo *pkginfo.Repo) {
	helpers.PrintBegin("fetching RPM repo from %s", repo.URI)
	err := download.RepoFiles(repo, u.Update)
	helpers.FailIfErr(err)
	helpers.PrintComplete("repo cached at %s", repo.RPMCache)
}

// ImportRepo stores the repo data from the cacheloc to the database
func ImportRepo(conf *config.Config, u *config.UInfo, repo *pkginfo.Repo) {
	helpers.PrintBegin("importing repo from %s to database", repo.RPMCache)
	err := pkginfo.ImportAllRPMs(repo, u.Update)
	helpers.FailIfErr(err)
	helpers.PrintComplete("RPM repo imported successfully")
}

// FetchRepo downloads RPMs from the repo.URI location and imports them into
// the database. This calls both the DownloadRepo and ImportRepo functions
func FetchRepo(conf *config.Config, u *config.UInfo) {
	repo, err := pkginfo.NewRepo(conf, u)
	helpers.FailIfErr(err)
	DownloadRepo(conf, u, &repo)
	ImportRepo(conf, u, &repo)
}

// DownloadBundles clones the bundle repository
func DownloadBundles(bundleInfo *pkginfo.BundleInfo) {
	helpers.PrintBegin("downloading bundle definitions")
	err := download.Bundles(bundleInfo)
	helpers.FailIfErr(err)
	helpers.PrintComplete("bundle repo cached to %s", bundleInfo.BundleCache)
}

// ImportBundles creates the bundle definitions, stores the current repo branch
// to revert back to after importing, and stores the bundle definitions in the
// database
func ImportBundles(bundleInfo *pkginfo.BundleInfo) {
	var err error

	helpers.PrintBegin("importing bundles from %s to database", bundleInfo.BundleCache)
	// checks out the correct version tag from bundles repository
	err = download.BundleVersion(bundleInfo)
	helpers.FailIfErr(err)
	err = pkginfo.ImportBundleDefinitions(bundleInfo)
	helpers.FailIfErr(err)
	helpers.PrintComplete("bundles imported successfully")

	// after fetching from the specified tag, defer back to previous state
	defer func() {
		_ = helpers.CheckoutBranch(bundleInfo.BundleCache, bundleInfo.Branch)
	}()
}

// FetchBundles clones the bundles repository from the config or passed in
// bundleURL argument and imports the information to the database.
// It calls the DownloadBundles and ImportBundles functions
func FetchBundles(conf *config.Config, u *config.UInfo) {
	bundleInfo, err := pkginfo.NewBundleInfo(conf, u)
	helpers.FailIfErr(err)

	DownloadBundles(&bundleInfo)
	ImportBundles(&bundleInfo)
}

// DownloadUpdate downloads the manifests from mInfo UpstreamURL to the
// minfo.Cacheloc
func DownloadUpdate(mInfo *pkginfo.ManifestInfo) {
	helpers.PrintBegin("Downloading manifests from %s at version %v", mInfo.UpstreamURL, mInfo.Version)
	err := download.UpdateContent(mInfo)
	helpers.FailIfErr(err)
	helpers.PrintComplete("manifests cached at %s", mInfo.CacheLoc)
}

// DownloadUpdateFiles downloads the manifests files from the upstreamURL to the
// mInfo.CacheLoc
func DownloadUpdateFiles(mInfo *pkginfo.ManifestInfo) {
	helpers.PrintBegin("Downloading manifests files from %s at version %v", mInfo.UpstreamURL, mInfo.Version)
	err := download.UpdateFiles(mInfo)
	helpers.FailIfErr(err)
	helpers.PrintComplete("manifest files cached at %s/update", mInfo.CacheLoc)
}

// DownloadUpdateAll downloads both the manifest and the manifest files to the
// cache location
func DownloadUpdateAll(mInfo *pkginfo.ManifestInfo) {
	DownloadUpdate(mInfo)
	DownloadUpdateFiles(mInfo)
}

// ImportUpdate imports the manifests from the cache location into the database
func ImportUpdate(mInfo *pkginfo.ManifestInfo) {
	helpers.PrintBegin("importing manifests from %s to database", mInfo.CacheLoc)
	err := pkginfo.ImportManifests(mInfo)
	helpers.FailIfErr(err)
	helpers.PrintComplete("manifest import complete")
}

// FetchUpdate downloads manifests from the u.URL server and TODO: imports them
// to the database
func FetchUpdate(conf *config.Config, u *config.UInfo) {
	mInfo, err := pkginfo.NewManifestInfo(conf, u)
	helpers.FailIfErr(err)
	DownloadUpdate(&mInfo)
	ImportUpdate(&mInfo)
}

// FetchUpdateFiles downloads relevant files for u.Ver from u.URL the update
// files are NOT stored in the database
func FetchUpdateFiles(conf *config.Config, u *config.UInfo) {
	mInfo, err := pkginfo.NewManifestInfo(conf, u)
	helpers.FailIfErr(err)
	DownloadUpdate(&mInfo)
	DownloadUpdateFiles(&mInfo)
}
