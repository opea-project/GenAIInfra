CHARTS_DIR := ./helm-charts

.PHONY: test

test:
	@for chart in $$(find $(CHARTS_DIR) -mindepth 1 -maxdepth 1 -type d); do \
		echo "Testing chart: $$chart"; \
		helm lint $$chart; \
	done

pre-commit:
	@pre-commit run --all-files