import { CardLogin } from "@/components/LoginCard";
import { createFileRoute } from "@tanstack/react-router";

function LoginPage(){
    return(
      <>
      <CardLogin />
      </>
    )
}

export const Route = createFileRoute("/login")({
  component: LoginPage,
});