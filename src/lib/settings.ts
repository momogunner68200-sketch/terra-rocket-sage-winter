import { create } from "zustand";
import { persist } from "zustand/middleware";

export type TextSize = "md" | "lg" | "xl";

type Settings = {
  size: TextSize;
  contrast: boolean;
  dwell: boolean;
  visual: boolean;
  setSize: (size: TextSize) => void;
  setContrast: (contrast: boolean) => void;
  setDwell: (dwell: boolean) => void;
  setVisual: (visual: boolean) => void;
};

export const useSettings = create<Settings>()(
  persist(
    (set) => ({
      size: "lg",
      contrast: false,
      dwell: false,
      visual: false,
      setSize: (size) => set({ size }),
      setContrast: (contrast) => set({ contrast }),
      setDwell: (dwell) => set({ dwell }),
      setVisual: (visual) => set({ visual }),
    }),
    { name: "portee-settings-v2", skipHydration: true },
  ),
);
