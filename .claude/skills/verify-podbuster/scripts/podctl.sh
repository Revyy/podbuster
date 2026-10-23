#!/usr/bin/env bash
# podctl: launch, inspect, drive and stop one Podbuster instance per run id.
# Every run lives in $PODBUSTER_VERIFY_ROOT/runs/<run-id>/ and survives stop.
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO="$(cd "${SKILL_DIR}/../../.." && pwd)"
ROOT="${PODBUSTER_VERIFY_ROOT:-${REPO}/.verify}"
KEYS="${ROOT}/bin/keys"
KIND_CONTEXT="kind-${KIND_CLUSTER_NAME:-podbuster-chaos}"
KIND_NAMESPACE="${PODBUSTER_NAMESPACE:-podbuster-chaos}"

die() { printf 'podctl: %s\n' "$*" >&2; exit 1; }
run_dir() { [[ -n "${1:-}" ]] || die "missing run id"; printf '%s/runs/%s' "${ROOT}" "$1"; }
pid_of() { cat "$(run_dir "$1")/pid" 2>/dev/null || die "run $1 was not launched"; }

# Our process: the pid file names a live process whose binary is this run's build.
alive() {
	local dir pid
	dir="$(run_dir "$1")"; pid="$(cat "${dir}/pid" 2>/dev/null)" || return 1
	[[ "$(ps -o comm= -p "${pid}" 2>/dev/null)" == "${dir}/podbuster" ]]
}

build_keys() {
	local src="${SKILL_DIR}/scripts/keys.swift"
	if [[ ! -x "${KEYS}" || "${src}" -nt "${KEYS}" ]]; then
		mkdir -p "${ROOT}/bin"
		swiftc -O "${src}" -o "${KEYS}"
	fi
}

# shot <run-id> <name>: press the Ebitengine screenshot key, move the PNG into shots/.
shot() {
	local id="$1" name="$2" dir pid before file i
	dir="$(run_dir "${id}")"; pid="$(pid_of "${id}")"
	alive "${id}" || die "run ${id} is not running"
	before="$(grep -c '^Saved screenshot: ' "${dir}/game.log" || true)"
	"${KEYS}" tap "${pid}" 60 f12
	for i in $(seq 1 50); do
		file="$(grep '^Saved screenshot: ' "${dir}/game.log" | sed -n "$((before + 1))p" | sed 's/^Saved screenshot: //' || true)"
		if [[ -n "${file}" && -s "${dir}/${file}" ]]; then
			mkdir -p "${dir}/shots"
			local n; n="$(ls "${dir}/shots" | wc -l | tr -d ' ')"
			local out; out="${dir}/shots/$(printf '%02d' "$((n + 1))")-${name}.png"
			mv "${dir}/${file}" "${out}"
			printf '%s\n' "${out}"
			return 0
		fi
		sleep 0.1
	done
	die "no screenshot appeared; run doctor ${id}"
}

cmd_launch() {
	local id="$1" mode="${2:-mock}" dir args
	dir="$(run_dir "${id}")"
	[[ -e "${dir}/pid" ]] && alive "${id}" && die "run ${id} is already running (pid $(pid_of "${id}"))"
	case "${mode}" in
		mock) args=(--mock --namespace verify) ;;
		kind)
			kubectl --context "${KIND_CONTEXT}" -n "${KIND_NAMESPACE}" get pods >/dev/null 2>&1 \
				|| die "kind playground not reachable; run 'make kind-up' first"
			args=(--context "${KIND_CONTEXT}" --namespace "${KIND_NAMESPACE}") ;;
		*) die "mode must be mock or kind (never a real cluster)" ;;
	esac
	command -v swiftc >/dev/null || die "swiftc is required (Xcode command line tools)"
	build_keys
	"${KEYS}" trusted || die "the host terminal lacks Accessibility permission; key events would be dropped"
	mkdir -p "${dir}/shots"
	(cd "${REPO}" && go build -o "${dir}/podbuster" .)
	{ printf 'mode=%s\nargs=%s\nrev=%s\n' "${mode}" "${args[*]}" "$(git -C "${REPO}" rev-parse --short HEAD)"
	  printf 'dirty=%s\n' "$(git -C "${REPO}" status --porcelain -- '*.go' go.mod go.sum | wc -l | tr -d ' ')"; } > "${dir}/run.env"
	(cd "${dir}"; EBITENGINE_SCREENSHOT_KEY=F12 nohup "${dir}/podbuster" "${args[@]}" > "${dir}/game.log" 2>&1 & echo $! > "${dir}/pid")
	local i
	for i in $(seq 1 100); do
		alive "${id}" || die "game exited during startup; see ${dir}/game.log"
		"${KEYS}" window "$(pid_of "${id}")" >/dev/null && break
		sleep 0.1
	done
	"${KEYS}" window "$(pid_of "${id}")" >/dev/null || die "no window after 10 s; see ${dir}/game.log"
	sleep 1  # let the first frames render
	"${KEYS}" activate "$(pid_of "${id}")"
	shot "${id}" ready >/dev/null
	printf 'ready: run %s pid %s dir %s\n' "${id}" "$(pid_of "${id}")" "${dir}"
}

cmd_doctor() {
	local id="$1" dir pid ok=0
	dir="$(run_dir "${id}")"; pid="$(pid_of "${id}")"
	check() { if eval "$2"; then printf 'ok    %s\n' "$1"; else printf 'FAIL  %s\n' "$1"; ok=1; fi; }
	check "process ${pid} is this run's build" 'alive "${id}"'
	check "window on screen" '"${KEYS}" window "${pid}" >/dev/null'
	check "Accessibility granted (key events delivered)" '"${KEYS}" trusted'
	check "build matches HEAD $(git -C "${REPO}" rev-parse --short HEAD)" 'grep -qx "rev=$(git -C "${REPO}" rev-parse --short HEAD)" "${dir}/run.env"'
	cat "${dir}/run.env"
	printf -- '--- last log lines (Metal "nextDrawable" lines at startup are harmless)\n'
	grep -v 'nextDrawable\|IMKClient\|IMKInputSession' "${dir}/game.log" | tail -5 || true
	return "${ok}"
}

cmd_key() {
	local id="$1" hold="$2"; shift 2
	alive "${id}" || die "run ${id} is not running"
	"${KEYS}" tap "$(pid_of "${id}")" "${hold}" "$@"
}

# spray <run-id> <taps>: turn left while firing the Blaster on every tap, a Rocket
# every 20th tap and a Barrel roll every 45th, to score kills without aiming.
cmd_spray() {
	local id="$1" taps="${2:-100}" pid i keys
	alive "${id}" || die "run ${id} is not running"
	pid="$(pid_of "${id}")"
	for i in $(seq 1 "${taps}"); do
		keys=(left space)
		(( i % 20 == 0 )) && keys+=(x)
		(( i % 45 == 0 )) && keys+=(z)
		"${KEYS}" tap "${pid}" 70 "${keys[@]}"
		sleep 0.05
	done
}

cmd_stop() {
	local id="$1"
	if alive "${id}"; then kill "$(pid_of "${id}")"; printf 'stopped run %s\n' "${id}"; else printf 'run %s not running\n' "${id}"; fi
	rm -f "$(run_dir "${id}")/podbuster"
	printf 'evidence kept in %s\n' "$(run_dir "${id}")"
}

case "${1:-}" in
	launch) shift; cmd_launch "$@" ;;
	doctor) shift; cmd_doctor "$@" ;;
	key) shift; cmd_key "$@" ;;
	spray) shift; cmd_spray "$@" ;;
	shot) shift; build_keys; shot "$@" ;;
	stop) shift; cmd_stop "$@" ;;
	*) die "usage: podctl.sh launch <run-id> [mock|kind] | doctor <run-id> | key <run-id> <holdMs> <key>... | spray <run-id> [taps] | shot <run-id> <name> | stop <run-id>" ;;
esac
