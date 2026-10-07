import { useEffect, useRef, useState } from "react";
import { Contrast, Eye } from "lucide-react";
import { clsx } from "clsx";
import { BigButton } from "@/components/big-button";
import { drawPractice, nextPrompt, readCenter, type QteKey } from "@/lib/center-read";
import { useSettings, type TextSize } from "@/lib/settings";

const SIZES: { id: TextSize; label: string }[] = [
  { id: "md", label: "A" },
  { id: "lg", label: "A+" },
  { id: "xl", label: "A++" },
];

const PAD: { key: QteKey; hint: string; place: string }[] = [
  { key: "Z", hint: "Haut", place: "col-start-2 row-start-1" },
  { key: "Q", hint: "Gauche", place: "col-start-1 row-start-2" },
  { key: "S", hint: "Bas", place: "col-start-2 row-start-2" },
  { key: "D", hint: "Droite", place: "col-start-3 row-start-2" },
];

export function PorteeApp() {
  const size = useSettings((s) => s.size);
  const contrast = useSettings((s) => s.contrast);
  const dwell = useSettings((s) => s.dwell);
  const visual = useSettings((s) => s.visual);
  const setSize = useSettings((s) => s.setSize);
  const setContrast = useSettings((s) => s.setContrast);
  const setDwell = useSettings((s) => s.setDwell);
  const setVisual = useSettings((s) => s.setVisual);

  const [mode, setMode] = useState<"trial" | "live">("trial");
  const [sharing, setSharing] = useState(false);
  const [prompt, setPrompt] = useState<QteKey>("Z");
  const [detected, setDetected] = useState<QteKey | null>(null);
  const [hits, setHits] = useState(0);
  const [wrong, setWrong] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const videoRef = useRef<HTMLVideoElement>(null);
  const stageRef = useRef<HTMLCanvasElement>(null);
  const previewRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const promptRef = useRef(prompt);
  const modeRef = useRef(mode);
  promptRef.current = prompt;
  modeRef.current = mode;

  useEffect(() => {
    void useSettings.persist.rehydrate();
  }, []);

  useEffect(() => {
    return () => {
      streamRef.current?.getTracks().forEach((track) => track.stop());
    };
  }, []);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;
    video.srcObject = streamRef.current;
  }, [sharing, mode]);

  useEffect(() => {
    const stage = stageRef.current;
    if (!stage || mode !== "trial") return;
    drawPractice(stage, prompt);
  }, [prompt, mode]);

  useEffect(() => {
    if (!visual || mode !== "trial" || !detected || detected !== prompt) return;
    press(detected);
  }, [visual, detected, prompt, mode]);

  useEffect(() => {
    let timer = 0;
    let stopped = false;
    const tick = () => {
      if (stopped) return;
      let found: QteKey | null = null;
      const preview = previewRef.current;
      if (modeRef.current === "live") {
        const video = videoRef.current;
        if (video && video.videoWidth > 0) {
          found = readCenter(video, video.videoWidth, video.videoHeight, preview);
        }
      } else {
        const stage = stageRef.current;
        if (stage && stage.width > 0) {
          found = readCenter(stage, stage.width, stage.height, preview);
        }
      }
      setDetected((current) => (current === found ? current : found));
      timer = window.setTimeout(tick, 320);
    };
    timer = window.setTimeout(tick, 200);
    return () => {
      stopped = true;
      window.clearTimeout(timer);
    };
  }, [mode, sharing, prompt]);

  function press(key: QteKey) {
    if (mode !== "trial") return;
    if (key === promptRef.current) {
      setWrong(false);
      setHits((count) => count + 1);
      setPrompt(nextPrompt(key));
      return;
    }
    setWrong(true);
    window.setTimeout(() => setWrong(false), 450);
  }

  async function shareScreen() {
    if (!navigator.mediaDevices?.getDisplayMedia) {
      setError("Ce navigateur ne permet pas le partage d'écran. L'entraînement reste disponible.");
      setMode("trial");
      return;
    }
    try {
      const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: false });
      stream.getAudioTracks().forEach((track) => {
        track.stop();
        stream.removeTrack(track);
      });
      streamRef.current?.getTracks().forEach((track) => track.stop());
      streamRef.current = stream;
      stream.getVideoTracks()[0]?.addEventListener("ended", () => {
        streamRef.current = null;
        setSharing(false);
        setMode("trial");
      });
      setSharing(true);
      setMode("live");
      setError(null);
    } catch {
      setError("Partage annulé. Aucune image n'est sortie de l'appareil.");
    }
  }

  function stopShare() {
    streamRef.current?.getTracks().forEach((track) => track.stop());
    streamRef.current = null;
    setSharing(false);
    setMode("trial");
  }

  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      const target = event.target as HTMLElement | null;
      if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement) return;
      const key = event.key.toUpperCase();
      if (key !== "Z" && key !== "Q" && key !== "S" && key !== "D") return;
      event.preventDefault();
      press(key);
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });

  return (
    <div
      data-size={size}
      data-contrast={contrast ? "high" : "normal"}
      className="flex h-dvh flex-col overflow-hidden bg-paper text-ink"
    >
      <a
        href="#scene"
        className="sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-3 focus:z-50 focus:rounded-lg focus:bg-card focus:px-4 focus:py-3"
      >
        Aller au centre
      </a>
      <div className="h-2 bg-action" />
      <header className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-3">
        <div className="flex items-center gap-3">
          <span className="grid size-11 shrink-0 place-items-center rounded-xl bg-action text-on-action" aria-hidden>
            <Eye className="size-6" />
          </span>
          <div>
            <p className="text-sm font-bold tracking-wide text-action uppercase">QTE local · sans micro</p>
            <h1 className="text-3xl font-bold tracking-tight">Portée</h1>
          </div>
        </div>
        <div className="flex flex-wrap gap-2" role="group" aria-label="Taille du texte">
          {SIZES.map((item) => (
            <button
              key={item.id}
              type="button"
              aria-pressed={size === item.id}
              onClick={() => setSize(item.id)}
              className={clsx(
                "min-h-12 min-w-12 rounded-xl border-2 px-3 font-bold",
                size === item.id ? "border-action bg-action text-on-action" : "border-line bg-card",
              )}
            >
              {item.label}
            </button>
          ))}
          <button
            type="button"
            aria-pressed={contrast}
            onClick={() => setContrast(!contrast)}
            className={clsx(
              "inline-flex min-h-12 items-center gap-2 rounded-xl border-2 px-4 font-bold",
              contrast ? "border-action bg-action text-on-action" : "border-line bg-card",
            )}
          >
            <Contrast className="size-5" aria-hidden />
            Contraste
          </button>
        </div>
      </header>

      <main className="mx-auto grid w-full max-w-6xl min-h-0 flex-1 gap-4 overflow-auto px-4 pb-3 lg:grid-cols-[minmax(0,1.3fr)_minmax(16rem,0.7fr)]">
        <section id="scene" aria-labelledby="scene-title" className="min-w-0">
          <div className="mb-3 flex flex-wrap items-end justify-between gap-3">
            <h2 id="scene-title" className="text-2xl font-bold">
              Centre de l'écran
            </h2>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                aria-pressed={mode === "trial"}
                onClick={() => setMode("trial")}
                className={clsx(
                  "min-h-12 rounded-xl border-2 px-3 font-bold",
                  mode === "trial" ? "border-action bg-action text-on-action" : "border-line bg-card",
                )}
              >
                Essai
              </button>
              <button
                type="button"
                aria-pressed={mode === "live"}
                onClick={() => {
                  if (sharing) setMode("live");
                  else void shareScreen();
                }}
                className={clsx(
                  "min-h-12 rounded-xl border-2 px-3 font-bold",
                  mode === "live" ? "border-action bg-action text-on-action" : "border-line bg-card",
                )}
              >
                Mon écran
              </button>
            </div>
          </div>

          {mode === "live" && sharing ? (
            <div className="relative overflow-hidden rounded-2xl border-2 border-ink bg-ink">
              <p className="bg-action px-4 py-3 font-bold text-on-action">
                Image seulement, sans son. Seul le cadre du milieu est lu, et il reste sur cet appareil.
              </p>
              <div className="relative">
                <video ref={videoRef} autoPlay muted playsInline className="max-h-80 w-full bg-ink" />
                <div className="pointer-events-none absolute top-1/2 left-1/2 h-1/4 -translate-x-1/2 -translate-y-1/2 border-4 border-action aspect-square" aria-hidden />
              </div>
            </div>
          ) : (
            <div className="relative overflow-hidden rounded-2xl border-2 border-ink bg-ink">
              <canvas ref={stageRef} className="h-auto w-full" aria-label={`Lettre d'entraînement ${prompt}`} />
              <div
                className="pointer-events-none absolute top-1/2 left-1/2 h-1/4 aspect-square -translate-x-1/2 -translate-y-1/2 border-4 border-action"
                aria-hidden
              />
            </div>
          )}
          {error ? (
            <p className="mt-3 rounded-xl border-2 border-ink p-3 font-bold" role="alert">
              {error}
            </p>
          ) : null}
          {sharing ? (
            <button type="button" className="mt-3 min-h-12 rounded-xl border-2 border-line px-4 font-bold" onClick={stopShare}>
              Arrêter le partage
            </button>
          ) : null}
        </section>

        <section aria-live="polite" className="min-w-0">
          <div className="rounded-2xl border-2 border-line bg-card p-4">
            <h2 className="text-2xl font-bold">Touche lue</h2>
            <p className="mt-2 text-pretty text-muted">
              Le carré au centre est agrandi ici. Quand une lettre Z, Q, S ou D y apparaît, elle s'affiche en grand.
            </p>
            <canvas ref={previewRef} className="mt-3 w-full rounded-xl border-2 border-line bg-ink" width={280} height={280} />
            <p className="mt-3 text-center text-6xl font-bold tabular-nums" data-testid="detected">
              {detected ?? "–"}
            </p>
            <p className={clsx("text-center font-bold", wrong ? "text-ink" : "text-muted")}>
              {mode === "trial"
                ? wrong
                  ? "Mauvaise touche. Regardez le centre."
                  : `Essai : ${hits} réussi${hits > 1 ? "s" : ""}`
                : "Dans le jeu, appuyez la touche affichée au clavier."}
            </p>
          </div>
          <label className="mt-3 flex min-h-12 items-start gap-3 rounded-xl border-2 border-line bg-card p-3 text-base font-bold">
            <input
              type="checkbox"
              className="mt-1 size-6 shrink-0 accent-action"
              checked={visual}
              onChange={(event) => setVisual(event.target.checked)}
            />
            <span>Souci visuel : effectuer le QTE automatiquement</span>
          </label>
          <p className="mt-3 text-pretty text-muted">
            {visual
              ? "Coché : la touche part seule quand la lettre est lue. Sur FiveM, Portee.exe fait la même chose dans la zone rouge."
              : "Cochez si la zone rouge est trop difficile à voir. Aucune touche ne part tant que c'est décoché."}
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            <a
              href="/Portee.zip"
              download
              className="inline-flex min-h-12 items-center rounded-xl border-2 border-action bg-action px-4 font-bold text-on-action"
            >
              Télécharger pour FiveM
            </a>
            <button
              type="button"
              aria-pressed={dwell}
              onClick={() => setDwell(!dwell)}
              className={clsx(
                "min-h-12 rounded-xl border-2 px-4 font-bold",
                dwell ? "border-action bg-action text-on-action" : "border-line bg-card",
              )}
            >
              Maintien {dwell ? "oui" : "non"}
            </button>
          </div>
        </section>
      </main>

      <div className="shrink-0 border-t-2 border-line bg-paper">
        <div className="mx-auto grid max-w-xl grid-cols-3 gap-2 px-4 py-3" role="group" aria-label="Touches Z Q S D">
          {PAD.map((item) => (
            <BigButton
              key={item.key}
              dwell={dwell}
              pressed={detected === item.key}
              onActivate={() => press(item.key)}
              className={item.place}
              ariaLabel={`${item.key}, ${item.hint}`}
              testId={`key-${item.key}`}
            >
              <span className="block text-3xl leading-none">{item.key}</span>
              <span className="mt-1 block text-sm">{item.hint}</span>
            </BigButton>
          ))}
        </div>
      </div>
    </div>
  );
}
