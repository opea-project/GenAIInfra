CHARTS_DIR := ./helm-charts

.PHONY: test

test:
	@for chart in $$(find $(CHARTS_DIR) -mindepth 1 -maxdepth 2 -type f -name "Chart.yaml"); do \
		  echo "Testing chart: $$(dirname $$chart)"; \
		  helm lint $$(dirname $$chart); \
	done

pre-commit:
	@pre-commit run --all-files
