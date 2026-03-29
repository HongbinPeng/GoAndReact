import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import store from "./store";
import { Provider } from "react-redux";
import { BrowserRouter, Routes, Route } from "react-router";
import Home from "./pages/Home";
import List from "./pages/List";
import About from "./pages/About";
import Login from "./pages/Login";
import Authorization from "./Authorization";

const root = createRoot(document.getElementById("root"));
root.render(
  <Provider store={store}>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />}>
          {/* 默认子路由 */}
          <Route index element={<Home />} />
          <Route
            path="list"
            element={
              <Authorization>
                <List />
              </Authorization>
            }
          />
          <Route
            path="about"
            element={
              <Authorization>
                <About />
              </Authorization>
            }
          />
        </Route>
        <Route path="login" element={<Login />} />
      </Routes>
    </BrowserRouter>
  </Provider>,
);

/* store.subscribe(() => {
  root.render(<App/>)
}) */
