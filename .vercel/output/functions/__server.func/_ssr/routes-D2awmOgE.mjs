import { i as __toESM } from "../_runtime.mjs";
import { b as require_jsx_runtime, q as require_react } from "../_libs/@tanstack/react-router+[...].mjs";
import { n as TSS_SERVER_FUNCTION, r as getServerFnById, t as createServerFn } from "./ssr.mjs";
import { n as Eye, r as Contrast } from "../_libs/lucide-react.mjs";
import { t as clsx } from "../_libs/clsx.mjs";
import { n as create, t as persist } from "../_libs/zustand.mjs";
//#region node_modules/.nitro/vite/services/ssr/assets/routes-D2awmOgE.js
var import_react = /* @__PURE__ */ __toESM(require_react());
var import_jsx_runtime = require_jsx_runtime();
var createSsrRpc = (functionId) => {
	const url = "/_serverFn/" + functionId;
	const serverFnMeta = { id: functionId };
	const fn = async (...args) => {
		return (await getServerFnById(functionId, { origin: "server" }))(...args);
	};
	return Object.assign(fn, {
		url,
		serverFnMeta,
		[TSS_SERVER_FUNCTION]: true
	});
};
var MAX_IMAGE_CHARS = 18e5;
var analyzeScreen = createServerFn({ method: "POST" }).validator((input) => {
	const image = input?.image;
	const source = input?.source;
	if (source !== "trial" && source !== "screen") throw new Error("Source de lecture inconnue.");
	if (typeof image !== "string" || !image.startsWith("data:image/jpeg;base64,")) throw new Error("Image refusée. Seule une photo JPEG de votre choix est acceptée.");
	if (image.length > MAX_IMAGE_CHARS) throw new Error("Image trop lourde. Rapprochez une fenêtre plus petite.");
	return {
		image,
		source
	};
}).handler(createSsrRpc("02d21f2e2cd255c18f85a105e5237fbab16ddca93f17ebc4f7e84d266c458623"));
var SERVICES = [
	{
		id: "etat-civil",
		title: "État civil",
		detail: "Naissance, mariage, décès"
	},
	{
		id: "identite",
		title: "Carte d'identité",
		detail: "Première demande ou renouvellement"
	},
	{
		id: "accessibilite",
		title: "Accès au bâtiment",
		detail: "Entrée, ascenseur, accueil"
	}
];
var initialTrial = {
	service: null,
	name: "",
	slot: "09:30",
	done: false
};
function serviceTitle(id) {
	return SERVICES.find((item) => item.id === id)?.title ?? "Aucun service";
}
function squeeze(canvas) {
	let quality = .72;
	let url = canvas.toDataURL("image/jpeg", quality);
	while (url.length > 14e5 && quality > .45) {
		quality -= .08;
		url = canvas.toDataURL("image/jpeg", quality);
	}
	if (url.length > 17e5) throw new Error("Image trop lourde. Partagez une fenêtre plus petite.");
	return url;
}
function captureVideo(video) {
	if (!video.videoWidth || !video.videoHeight) throw new Error("L'écran partagé n'est pas encore prêt. Attendez une seconde, puis réessayez.");
	const scale = Math.min(1, 960 / video.videoWidth);
	const canvas = document.createElement("canvas");
	canvas.width = Math.max(1, Math.round(video.videoWidth * scale));
	canvas.height = Math.max(1, Math.round(video.videoHeight * scale));
	const ctx = canvas.getContext("2d");
	if (!ctx) throw new Error("Impossible de photographier l'écran.");
	ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
	return squeeze(canvas);
}
function paintTrial(state) {
	const canvas = document.createElement("canvas");
	canvas.width = 960;
	canvas.height = 720;
	const ctx = canvas.getContext("2d");
	if (!ctx) throw new Error("Impossible de photographier la scène d'essai.");
	ctx.fillStyle = "#f6f1e7";
	ctx.fillRect(0, 0, 960, 720);
	ctx.fillStyle = "#fffdf9";
	ctx.strokeStyle = "#ddd4c6";
	ctx.lineWidth = 2;
	ctx.beginPath();
	ctx.roundRect(40, 36, 880, 648, 24);
	ctx.fill();
	ctx.stroke();
	ctx.fillStyle = "#0e6b58";
	ctx.font = "700 26px sans-serif";
	ctx.fillText("Guichet Clair", 72, 96);
	ctx.fillStyle = "#1c1916";
	ctx.font = "700 40px sans-serif";
	ctx.fillText("Prendre rendez-vous", 72, 156);
	ctx.fillStyle = "#5c564c";
	ctx.font = "22px sans-serif";
	ctx.fillText("Scène d'essai — guichet fictif, rien n'est envoyé à une mairie.", 72, 198);
	SERVICES.forEach((service, index) => {
		const x = 72 + index * 276;
		const selected = state.service === service.id;
		ctx.beginPath();
		ctx.roundRect(x, 236, 256, 128, 16);
		ctx.fillStyle = selected ? "#0e6b58" : "#fffdf9";
		ctx.fill();
		ctx.lineWidth = 3;
		ctx.strokeStyle = selected ? "#0e6b58" : "#1c1916";
		ctx.stroke();
		ctx.fillStyle = selected ? "#f6f1e7" : "#1c1916";
		ctx.font = "700 24px sans-serif";
		ctx.fillText(service.title, x + 18, 292);
		ctx.font = "18px sans-serif";
		wrapText(ctx, service.detail, x + 18, 328, 220, 24);
	});
	ctx.fillStyle = "#1c1916";
	if (state.done) {
		ctx.font = "700 30px sans-serif";
		const who = state.name.trim().slice(0, 42) || "sans nom indiqué";
		ctx.fillText(`Rendez-vous noté pour ${who}`, 72, 450);
		ctx.font = "24px sans-serif";
		ctx.fillText(`Créneau ${state.slot} · ${serviceTitle(state.service)}`, 72, 494);
		ctx.fillStyle = "#5c564c";
		ctx.font = "20px sans-serif";
		ctx.fillText("Bouton visible : Nouveau rendez-vous", 72, 548);
	} else if (state.service) {
		ctx.font = "700 24px sans-serif";
		ctx.fillText("Votre nom", 72, 430);
		ctx.strokeStyle = "#1c1916";
		ctx.lineWidth = 3;
		ctx.strokeRect(72, 448, 520, 64);
		ctx.font = "24px sans-serif";
		ctx.fillText(state.name.trim().slice(0, 42) || "(vide)", 88, 490);
		ctx.fillText(`Créneau choisi : ${state.slot}`, 72, 560);
		ctx.fillStyle = "#0e6b58";
		ctx.beginPath();
		ctx.roundRect(72, 588, 260, 64, 12);
		ctx.fill();
		ctx.fillStyle = "#f6f1e7";
		ctx.font = "700 24px sans-serif";
		ctx.fillText("Confirmer", 138, 628);
	} else {
		ctx.font = "22px sans-serif";
		ctx.fillText("Aucun service choisi. Les trois cartes sont les boutons.", 72, 430);
	}
	return squeeze(canvas);
}
function wrapText(ctx, text, x, y, maxWidth, lineHeight) {
	const words = text.split(" ");
	let line = "";
	let cursor = y;
	for (const word of words) {
		const next = line ? `${line} ${word}` : word;
		if (ctx.measureText(next).width > maxWidth && line) {
			ctx.fillText(line, x, cursor);
			line = word;
			cursor += lineHeight;
		} else line = next;
	}
	if (line) ctx.fillText(line, x, cursor);
}
var useSettings = create()(persist((set) => ({
	size: "lg",
	contrast: false,
	dwell: false,
	setSize: (size) => set({ size }),
	setContrast: (contrast) => set({ contrast }),
	setDwell: (dwell) => set({ dwell })
}), {
	name: "portee-settings-v2",
	skipHydration: true
}));
function BigButton({ children, onActivate, dwell, dwellMs = 1100, variant = "ghost", disabled = false, pressed, ariaLabel, className, testId }) {
	const [progress, setProgress] = (0, import_react.useState)(0);
	const frame = (0, import_react.useRef)(0);
	const fired = (0, import_react.useRef)(false);
	function holdStart() {
		if (!dwell || disabled) return;
		fired.current = false;
		const started = performance.now();
		const step = (now) => {
			const ratio = Math.min(1, (now - started) / dwellMs);
			setProgress(ratio);
			if (ratio >= 1) {
				cancelAnimationFrame(frame.current);
				setProgress(0);
				fired.current = true;
				onActivate();
				return;
			}
			frame.current = requestAnimationFrame(step);
		};
		cancelAnimationFrame(frame.current);
		frame.current = requestAnimationFrame(step);
	}
	function holdEnd() {
		cancelAnimationFrame(frame.current);
		setProgress(0);
	}
	return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("button", {
		type: "button",
		"data-testid": testId,
		disabled,
		"aria-label": ariaLabel,
		"aria-pressed": pressed,
		onClick: () => {
			if (fired.current) {
				fired.current = false;
				return;
			}
			onActivate();
		},
		onPointerDown: holdStart,
		onPointerUp: holdEnd,
		onPointerLeave: holdEnd,
		onPointerCancel: holdEnd,
		className: clsx("relative isolate min-h-14 overflow-hidden rounded-xl border-2 px-4 py-2 text-center text-lg font-bold leading-tight", "disabled:cursor-not-allowed disabled:opacity-50", variant === "action" ? "border-action bg-action text-on-action" : "border-line bg-card text-ink", className),
		children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
			className: "relative z-10",
			children
		}), dwell && progress > 0 ? /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
			className: "absolute inset-x-0 bottom-0 z-20 h-1.5 bg-ink/20",
			"aria-hidden": true,
			children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
				className: "block h-full bg-action",
				style: { width: `${Math.round(progress * 100)}%` }
			})
		}) : null]
	});
}
function TrialDesk({ state, onChange }) {
	return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
		className: "rounded-2xl border-2 border-line bg-card p-4 sm:p-6",
		children: [
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
				className: "text-sm font-bold tracking-wide text-action uppercase",
				children: "Guichet Clair"
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h3", {
				className: "mt-2 text-2xl font-bold text-balance",
				children: "Prendre rendez-vous"
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
				className: "mt-2 max-w-xl text-pretty text-muted",
				children: "Scène d'essai. Ce guichet est fictif : rien n'est envoyé à une administration."
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
				className: "mt-5 grid gap-3 sm:grid-cols-3",
				role: "group",
				"aria-label": "Choisir un service",
				children: SERVICES.map((service) => {
					const selected = state.service === service.id;
					return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("button", {
						type: "button",
						"aria-pressed": selected,
						onClick: () => onChange({
							...state,
							service: service.id,
							done: false
						}),
						className: clsx("min-h-24 rounded-xl border-2 px-3 py-3 text-left", selected ? "border-action bg-action text-on-action" : "border-ink bg-card text-ink"),
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
							className: "block font-bold",
							children: service.title
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
							className: clsx("mt-1 block", selected ? "text-on-action" : "text-muted"),
							children: service.detail
						})]
					}, service.id);
				})
			}),
			state.done ? /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
				className: "mt-5 rounded-xl border-2 border-action p-4",
				role: "status",
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("p", {
						className: "font-bold",
						children: [
							"Rendez-vous noté",
							state.name.trim() ? ` pour ${state.name.trim()}` : "",
							"."
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("p", {
						className: "mt-1",
						children: [
							serviceTitle(state.service),
							" · créneau ",
							state.slot,
							"."
						]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
						type: "button",
						className: "mt-4 min-h-14 rounded-xl border-2 border-ink bg-card px-4 font-bold",
						onClick: () => onChange({
							service: null,
							name: "",
							slot: "09:30",
							done: false
						}),
						children: "Nouveau rendez-vous"
					})
				]
			}) : state.service ? /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("form", {
				className: "mt-5 grid gap-4",
				onSubmit: (event) => {
					event.preventDefault();
					onChange({
						...state,
						done: true
					});
				},
				children: [
					/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("label", {
						className: "grid gap-2 font-bold",
						htmlFor: "trial-name",
						children: ["Votre nom", /* @__PURE__ */ (0, import_jsx_runtime.jsx)("input", {
							id: "trial-name",
							value: state.name,
							autoComplete: "name",
							onChange: (event) => onChange({
								...state,
								name: event.target.value.slice(0, 60)
							}),
							className: "min-h-14 rounded-xl border-2 border-ink bg-card px-3 font-normal text-ink"
						})]
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
						role: "group",
						"aria-label": "Créneau",
						className: "grid grid-cols-2 gap-3",
						children: ["09:30", "14:15"].map((slot) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
							type: "button",
							"aria-pressed": state.slot === slot,
							onClick: () => onChange({
								...state,
								slot
							}),
							className: clsx("min-h-14 rounded-xl border-2 font-bold", state.slot === slot ? "border-action bg-action text-on-action" : "border-line bg-card text-ink"),
							children: slot
						}, slot))
					}),
					/* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
						type: "submit",
						className: "min-h-16 rounded-xl border-2 border-action bg-action px-4 font-bold text-on-action",
						children: "Confirmer"
					})
				]
			}) : /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
				className: "mt-5 text-muted",
				children: "Choisissez un service pour continuer."
			})
		]
	});
}
var SIZES = [
	{
		id: "md",
		label: "A"
	},
	{
		id: "lg",
		label: "A+"
	},
	{
		id: "xl",
		label: "A++"
	}
];
function PorteeApp() {
	const size = useSettings((s) => s.size);
	const contrast = useSettings((s) => s.contrast);
	const dwell = useSettings((s) => s.dwell);
	const setSize = useSettings((s) => s.setSize);
	const setContrast = useSettings((s) => s.setContrast);
	const setDwell = useSettings((s) => s.setDwell);
	const [mode, setMode] = (0, import_react.useState)("trial");
	const [trial, setTrial] = (0, import_react.useState)(initialTrial);
	const [reading, setReading] = (0, import_react.useState)(null);
	const [status, setStatus] = (0, import_react.useState)("idle");
	const [error, setError] = (0, import_react.useState)(null);
	const [sharing, setSharing] = (0, import_react.useState)(false);
	const [zoom, setZoom] = (0, import_react.useState)(1);
	const videoRef = (0, import_react.useRef)(null);
	const streamRef = (0, import_react.useRef)(null);
	const busyRef = (0, import_react.useRef)(false);
	(0, import_react.useEffect)(() => {
		useSettings.persist.rehydrate();
	}, []);
	(0, import_react.useEffect)(() => {
		return () => {
			streamRef.current?.getTracks().forEach((track) => track.stop());
		};
	}, []);
	(0, import_react.useEffect)(() => {
		const video = videoRef.current;
		if (!video) return;
		video.srcObject = streamRef.current;
	}, [sharing, mode]);
	async function runAnalyze() {
		if (busyRef.current) return;
		busyRef.current = true;
		setStatus("working");
		setError(null);
		try {
			const video = videoRef.current;
			const live = mode === "live" && Boolean(video && video.videoWidth > 0);
			if (mode === "live" && !live) {
				setStatus("error");
				setError("L'écran partagé n'est pas prêt. Attendez l'image, ou revenez à l'essai.");
				return;
			}
			const result = await analyzeScreen({ data: {
				image: live && video ? captureVideo(video) : paintTrial(trial),
				source: live ? "screen" : "trial"
			} });
			if (!result.ok) {
				setStatus("error");
				setError(result.error);
				return;
			}
			setReading(result.reading);
			setStatus("idle");
		} catch (cause) {
			setStatus("error");
			setError(cause instanceof Error ? cause.message : "Lecture impossible.");
		} finally {
			busyRef.current = false;
		}
	}
	async function shareScreen() {
		if (!navigator.mediaDevices?.getDisplayMedia) {
			setError("Ce navigateur ne permet pas le partage d'écran. La scène d'essai reste disponible.");
			setMode("trial");
			return;
		}
		try {
			const stream = await navigator.mediaDevices.getDisplayMedia({
				video: true,
				audio: false
			});
			stream.getAudioTracks().forEach((track) => {
				track.stop();
				stream.removeTrack(track);
			});
			streamRef.current?.getTracks().forEach((track) => track.stop());
			streamRef.current = stream;
			const [track] = stream.getVideoTracks();
			track?.addEventListener("ended", () => {
				streamRef.current = null;
				setSharing(false);
				setMode("trial");
			});
			setSharing(true);
			setMode("live");
			setZoom(1);
			setError(null);
		} catch {
			setError("Partage annulé. Rien n'a été vu.");
		}
	}
	const analyzeRef = (0, import_react.useRef)(runAnalyze);
	analyzeRef.current = runAnalyze;
	(0, import_react.useEffect)(() => {
		function onKey(event) {
			const target = event.target;
			if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target?.isContentEditable) return;
			if (event.key === "a" || event.key === "A") {
				event.preventDefault();
				analyzeRef.current();
			}
		}
		window.addEventListener("keydown", onKey);
		return () => window.removeEventListener("keydown", onKey);
	}, []);
	const working = status === "working";
	return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
		"data-size": size,
		"data-contrast": contrast ? "high" : "normal",
		className: "flex h-dvh flex-col overflow-hidden bg-paper text-ink",
		children: [
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("a", {
				href: "#scene",
				className: "sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-3 focus:z-50 focus:rounded-lg focus:bg-card focus:px-4 focus:py-3",
				children: "Aller à l'écran"
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", { className: "h-2 bg-action" }),
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("header", {
				className: "mx-auto flex w-full max-w-6xl flex-col gap-3 px-4 py-4",
				children: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
					className: "flex flex-wrap items-center justify-between gap-3",
					children: [/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "flex items-center gap-3",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
							className: "grid size-12 shrink-0 place-items-center rounded-xl bg-action text-on-action",
							"aria-hidden": true,
							children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)(Eye, { className: "size-6" })
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", { children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "text-sm font-bold tracking-wide text-action uppercase",
							children: "Aide à l'écran · sans micro"
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("h1", {
							className: "text-3xl font-bold tracking-tight text-balance sm:text-4xl",
							children: "Portée"
						})] })]
					}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "flex flex-wrap gap-2",
						role: "group",
						"aria-label": "Taille du texte",
						children: [SIZES.map((item) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
							type: "button",
							"aria-pressed": size === item.id,
							onClick: () => setSize(item.id),
							className: clsx("min-h-12 min-w-12 rounded-xl border-2 px-3 font-bold", size === item.id ? "border-action bg-action text-on-action" : "border-line bg-card"),
							children: item.label
						}, item.id)), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("button", {
							type: "button",
							"aria-pressed": contrast,
							onClick: () => setContrast(!contrast),
							className: clsx("inline-flex min-h-12 items-center gap-2 rounded-xl border-2 px-4 font-bold", contrast ? "border-action bg-action text-on-action" : "border-line bg-card"),
							children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)(Contrast, {
								className: "size-5",
								"aria-hidden": true
							}), "Contraste"]
						})]
					})]
				})
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("main", {
				className: "mx-auto grid w-full max-w-6xl min-h-0 flex-1 gap-6 overflow-auto px-4 py-2 lg:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.9fr)]",
				children: [/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
					id: "scene",
					"aria-labelledby": "scene-title",
					className: "min-w-0",
					children: [/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "mb-3 flex flex-wrap items-end justify-between gap-3",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h2", {
							id: "scene-title",
							className: "text-2xl font-bold",
							children: "Écran"
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
							className: "grid grid-cols-2 gap-2",
							role: "group",
							"aria-label": "Source",
							children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
								type: "button",
								"aria-pressed": mode === "trial",
								onClick: () => setMode("trial"),
								className: clsx("min-h-12 rounded-xl border-2 px-3 font-bold", mode === "trial" ? "border-action bg-action text-on-action" : "border-line bg-card"),
								children: "Essai"
							}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
								type: "button",
								"aria-pressed": mode === "live",
								onClick: () => {
									if (sharing) setMode("live");
									else shareScreen();
								},
								className: clsx("min-h-12 rounded-xl border-2 px-3 font-bold", mode === "live" ? "border-action bg-action text-on-action" : "border-line bg-card"),
								children: "Mon écran"
							})]
						})]
					}), mode === "live" && sharing ? /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "overflow-hidden rounded-2xl border-2 border-ink bg-ink",
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "bg-action px-4 py-3 font-bold text-on-action",
								children: "Partage en cours, image seulement, sans micro. Rien n'est analysé tant que vous n'appuyez pas sur Analyser."
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
								className: "max-h-[28rem] overflow-hidden",
								children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("video", {
									ref: videoRef,
									autoPlay: true,
									muted: true,
									playsInline: true,
									className: "w-full origin-center bg-ink",
									style: { transform: `scale(${zoom})` }
								})
							}),
							/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "flex gap-2 bg-card p-3",
								children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
									type: "button",
									className: "min-h-12 flex-1 rounded-xl border-2 border-line font-bold",
									onClick: () => setZoom((value) => Math.max(1, Number((value - .25).toFixed(2)))),
									children: "Loupe −"
								}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("button", {
									type: "button",
									className: "min-h-12 flex-1 rounded-xl border-2 border-line font-bold",
									onClick: () => setZoom((value) => Math.min(2.5, Number((value + .25).toFixed(2)))),
									children: "Loupe +"
								})]
							})
						]
					}) : /* @__PURE__ */ (0, import_jsx_runtime.jsx)(TrialDesk, {
						state: trial,
						onChange: setTrial
					})]
				}), /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("section", {
					"aria-labelledby": "guide-title",
					"aria-live": "polite",
					className: "min-w-0",
					children: [/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "rounded-2xl border-2 border-line bg-card p-4 sm:p-5",
						children: [
							/* @__PURE__ */ (0, import_jsx_runtime.jsx)("h2", {
								id: "guide-title",
								className: "text-2xl font-bold",
								children: "Guide"
							}),
							working ? /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "mt-4 font-bold",
								role: "status",
								children: "Lecture en cours…"
							}) : null,
							error ? /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "mt-4 rounded-xl border-2 border-ink p-3 font-bold",
								role: "alert",
								children: error
							}) : null,
							!reading && !working && !error ? /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
								className: "mt-4 text-pretty leading-relaxed text-muted",
								children: "Choisissez la scène d'essai ou partagez votre écran, puis appuyez sur Analyser. Le texte indique ce qui est visible, où agir, et ce qui est difficile à cliquer. Aucun micro."
							}) : null,
							reading ? /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
								className: "mt-4 grid gap-4",
								children: [
									/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
										className: "text-pretty leading-relaxed",
										children: reading.summary
									}),
									reading.readAloud && reading.readAloud !== reading.summary ? /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
										className: "text-pretty leading-relaxed",
										children: reading.readAloud
									}) : null,
									reading.nextStep ? /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
										className: "rounded-xl border-2 border-action p-3",
										children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
											className: "text-sm font-bold tracking-wide text-action uppercase",
											children: "Prochaine étape"
										}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
											className: "mt-1 font-bold text-pretty",
											children: reading.nextStep
										})]
									}) : null,
									reading.elements.length > 0 ? /* @__PURE__ */ (0, import_jsx_runtime.jsx)("ul", {
										className: "grid gap-3",
										children: reading.elements.map((item) => /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("li", {
											className: "border-t border-line pt-3",
											children: [
												/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
													className: "font-bold",
													children: item.name
												}),
												/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
													className: "text-muted",
													children: item.place
												}),
												/* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
													className: "mt-1 text-pretty",
													children: item.action
												})
											]
										}, `${item.name}-${item.place}`))
									}) : null,
									reading.barriers ? /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("p", {
										className: "text-pretty",
										children: [/* @__PURE__ */ (0, import_jsx_runtime.jsx)("span", {
											className: "font-bold",
											children: "Obstacle possible. "
										}), reading.barriers]
									}) : null
								]
							}) : null
						]
					}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
						className: "mt-4 grid gap-3",
						children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("p", {
							className: "text-pretty text-muted",
							children: "Maintien : gardez le doigt environ une seconde. Clavier : A pour analyser. Aucun micro n'est demandé, et le partage d'écran est sans son. L'image n'est pas conservée."
						})
					})]
				})]
			}),
			/* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
				className: "z-20 shrink-0 border-t-2 border-line bg-paper",
				children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)("div", {
					className: "mx-auto flex max-w-6xl flex-col gap-2 px-4 py-3 sm:flex-row sm:items-center",
					children: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)("div", {
						className: "grid w-full gap-2 sm:grid-cols-[auto_minmax(0,1fr)] sm:items-center",
						children: [/* @__PURE__ */ (0, import_jsx_runtime.jsxs)("button", {
							type: "button",
							"aria-pressed": dwell,
							onClick: () => setDwell(!dwell),
							className: clsx("min-h-14 rounded-xl border-2 px-3 font-bold", dwell ? "border-action bg-action text-on-action" : "border-line bg-card"),
							children: ["Maintien ", dwell ? "oui" : "non"]
						}), /* @__PURE__ */ (0, import_jsx_runtime.jsx)(BigButton, {
							variant: "action",
							dwell,
							disabled: working,
							testId: "analyze",
							className: "sm:flex-1",
							onActivate: () => void runAnalyze(),
							children: working ? "Lecture…" : "Analyser"
						})]
					})
				})
			})
		]
	});
}
function Home() {
	return /* @__PURE__ */ (0, import_jsx_runtime.jsx)(PorteeApp, {});
}
//#endregion
export { Home as component };
