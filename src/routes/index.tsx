import { createFileRoute } from "@tanstack/react-router";
import { PorteeApp } from "@/components/portee-app";

export const Route = createFileRoute("/")({ component: Home });

function Home() {
  return <PorteeApp />;
}
