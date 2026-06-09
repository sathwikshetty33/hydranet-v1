import { useEffect } from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import Nav from './components/layout/Nav';
import Footer from './components/layout/Footer';
import Landing from './pages/Landing';
import Agents from './pages/Agents';

function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);
  return null;
}

export default function App() {
  return (
    <>
      <ScrollToTop />
      <Nav />
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/agents" element={<Agents />} />
      </Routes>
      <Footer />
    </>
  );
}
