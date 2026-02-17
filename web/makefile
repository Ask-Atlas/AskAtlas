dev:
	infisical run --env=dev -- pnpm run dev

staging:
	infisical run --env=staging -- sh -c "pnpm run build && pnpm run start"

prod:
	infisical run --env=prod -- sh -c "pnpm run build && pnpm run start"

build:
	pnpm run build

install:
	pnpm install

clean:
	rm -rf .next node_modules

format:
	pnpm exec prettier --write "**/*.{ts,tsx,js,jsx,json,css,md}"

format-check:
	pnpm exec prettier --check "**/*.{ts,tsx,js,jsx,json,css,md}"

lint:
	pnpm run lint

lint-fix:
	pnpm run lint --fix

typecheck:
	pnpm exec tsc --noEmit

test:
	pnpm run test

test-e2e:
	infisical run --env=dev -- pnpm exec playwright test

test-e2e-staging:
	infisical run --env=staging -- pnpm exec playwright test

test-e2e-prod:
	infisical run --env=prod -- pnpm exec playwright test

test-e2e-codegen:
	infisical run --env=dev -- pnpm exec playwright codegen http://localhost:3000

e2e-report:
	pnpm exec playwright show-report

docker-build:
	docker build -t askatlas-web .

docker-run-local:
	docker run -p 3000:3000 --env-file .env.local askatlas-web

.PHONY: dev build staging prod clean install format lint lint-fix format-check test typecheck test-e2e test-e2e-staging test-e2e-prod test-e2e-codegen e2e-report docker-build docker-run-local