import { Navigate, useLocation } from "react-router";

export default function Authorization({ children }) {
  const whiteList = ["/", "/about"];

  const { pathname } = useLocation();

  const notRequiredAuth = whiteList.includes(pathname);

  const hasToken = false;
  return hasToken || notRequiredAuth ? children : <Navigate to="/login" replace state={pathname} />;
}