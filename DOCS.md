## Documentation

The OpenFaaS docs website is generated from Markdown in the `docs` subdirectory using mkdocs.

Start a development server in Docker by running the following command in the root of your `faas` workspace:

    docker run --rm -it -p 8000:8000 -v `pwd`:/docs squidfunk/mkdocs-material

This will start the development server running on http://localhost:8000, any changes made to the docs in your workspace will be picked up automatically and the site regenerated live.