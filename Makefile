.PHONY: replace-module

replace-module:
	@read -p "Enter the directory path where the replacement exists: " dir; \
	if [ ! -d "$$dir" ]; then \
		echo "\033[0;31mError: Invalid directory path\033[0m"; \
		exit 1; \
	fi; \
	if [ ! -f "$$dir/go.mod" ]; then \
		echo "\033[0;31mError: go.mod does not exist in the provided directory\033[0m"; \
		exit 1; \
	fi; \
	module_name=$$(grep -E "^module\s+" $$dir/go.mod | awk '{print $$2}'); \
	if [ -z "$$module_name" ]; then \
		echo "\033[0;31mError: Unable to extract module name from go.mod\033[0m"; \
		exit 1; \
	fi; \
	cmd="go mod edit -replace=$$module_name=$$dir"; \
	echo "\033[0;36m$$cmd\033[0m"; \
	GOPRIVATE=$$module_name $$cmd; \
	if ! grep -qE "replace\s+$$module_name\s+=>\s+$$dir" go.mod; then \
		echo "\033[0;31mError: Replacement not found in go.mod\033[0m"; \
		exit 1; \
	fi; \
	go clean -modcache; \
	echo "\033[0;32mReplacement added successfully and module cache cleaned!\033[0m"; \
	echo "\033[0;33m\nTo set the module as private system-wide, follow one of the following methods:\033[0m"; \
	echo "1. Pass the GOPRIVATE env var to the go binary: \033[0;34mGOPRIVATE=$$module_name go [command]\033[0m"; \
	echo "2. Add the following to your ~/.bashrc (or equivalent for your shell):"; \
	echo "   \033[0;34mexport GOPRIVATE=$$module_name\033[0m"; \
	echo "   Then run: \033[0;34msource ~/.bashrc\033[0m"; \
	echo "3. Or simply run the following in your current shell:"; \
	echo "   \033[0;34mexport GOPRIVATE=$$module_name\033[0m";
