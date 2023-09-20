.PHONY: replace-module clean-replace clean-module_name

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
	original_replace=$$(grep -E "replace\s+$$module_name\s+=>" go.mod); \
	if [ "$$original_replace" ]; then \
		if [ ! -f .original_replace ]; then \
			echo "$$original_replace" > .original_replace; \
		fi; \
		echo "\033[0;33mA replace directive for $$module_name already exists in go.mod. Saving the original directive.\033[0m"; \
		echo "\033[0;33mAdding a git pre-commit hook to restore the original replace directive when staged.\033[0m"; \
		if [ ! -f .git/hooks/pre-commit ]; then \
			echo '#!/bin/sh' > .git/hooks/pre-commit; \
			chmod +x .git/hooks/pre-commit; \
		fi; \
		MARKER="# REPLACE MODULE HOOK"; \
		HOOK_CONTENT='\
		if git diff --cached --name-only | grep -q "go.mod"; then \
			original_replace=$$(cat .original_replace); \
			modified_replace=$$(grep -E "replace\s+'$$module_name'\s+=>" go.mod); \
			sed -i "s|$$modified_replace|$$original_replace|" go.mod; \
			go_version=$$(grep "^go [0-9]\+\.[0-9]\+" go.mod | awk '\''{split($$0, a, " "); print a[2]}'\''); \
			echo "go mod tidy -compat=$$go_version"; \
			go mod tidy -compat=$$go_version; \
			git add go.mod go.sum; \
		fi'; \
		if ! grep -q "$$MARKER" .git/hooks/pre-commit; then \
			echo "$$MARKER" >> .git/hooks/pre-commit; \
			echo "$$HOOK_CONTENT" >> .git/hooks/pre-commit; \
			echo "# END REPLACE MODULE HOOK" >> .git/hooks/pre-commit; \
		fi; \
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

clean-replace:
	@if [ -f .original_replace ]; then \
		rm .original_replace; \
		echo "\033[0;32mRemoved .original_replace\033[0m"; \
	fi; \
	if [ -f .git/hooks/pre-commit ]; then \
		sed -i '/# REPLACE MODULE HOOK/,/# END REPLACE MODULE HOOK/d' .git/hooks/pre-commit; \
		echo "\033[0;32mRemoved replace module hook from .git/hooks/pre-commit\033[0m"; \
	fi;

	
clean-module:
	@go_version=$$(grep "^go [0-9]\+\.[0-9]\+" go.mod | cut -d' ' -f2); \
	go clean -modcache; \
	go mod tidy -compat=$$go_version;