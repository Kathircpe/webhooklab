import { Routes, Route, Link } from "react-router-dom";
import EndpointsPage from "./pages/EndpointsPage";
import EndpointDetailPage from "./pages/EndpointDetailPage";

export default function App() {
  return (
    <div className="container">
      <header><h1><Link to="/">WebhookLab</Link></h1></header>
      <Routes>
        <Route path="/" element={<EndpointsPage />} />
        <Route path="/endpoints/:id" element={<EndpointDetailPage />} />
      </Routes>
    </div>
  );
}
