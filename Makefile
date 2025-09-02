gen-ref:
	scripts/gen_ref_docs.sh

serve-docs:
	uvx -with mkdocs-simple-blog --from mkdocs-material mkdocs serve

build-docs:
	uvx -with mkdocs-simple-blog --from mkdocs-material mkdocs build --strict
