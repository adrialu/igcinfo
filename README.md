# IGC Track Viewer

Created as part of an assignment in Cloud Technologies (IMT2681) at NTNU Gjøvik 2018.

This application is a web service that allow users to upload and browse information about IGC files.  
IGC is an international file format for soaring track files that are used by (para)gliders.

A demo is available at <https://igcinfo-adrialu.herokuapp.com>.

It's an RESTful API with 5 available calls:

- `GET /api`
	Returns system status.
- `GET /api/igc`
	Returns an array of all stored tracks.
- `POST /api/igc`
	Takes `{"url":"<url>"}` as JSON data and returns the assigned ID.
- `GET /api/igc/<id>`
	Returns track data (fields) for a valid `id`.
- `GET /api/igc/<id>/<field>`
	Returns singular values for a `field` for a valid `id`.

## Notes

No particular coding style, linting or documentation has been enforced in this application,
and very few external packages has been used.

Further development would include linting, tests and use of well-established packages, but none of
these were a requirement/encouraged for the assignment.

[Browse back to 1f7a7408c3872987b244c67bbcd162479eb8b0fe](https://github.com/adrialu/igcinfo/tree/1f7a7408c3872987b244c67bbcd162479eb8b0fe) to see the implementation done without http routing/rendering packages.
