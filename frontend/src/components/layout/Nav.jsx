import { useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import './Nav.css';

export default function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const { pathname } = useLocation();

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  return (
    <header className={`nav ${scrolled ? 'nav--scrolled' : ''}`}>
      <div className="nav__inner shell">
        <Link to="/" className="nav__brand">
          <span className="nav__glyph" aria-hidden="true" />
          <span className="nav__word t-display">Hydranet</span>
        </Link>

        <nav className="nav__links">
          <a href="/#how" className="nav__link t-mono">Pipeline</a>
          <a href="/#capabilities" className="nav__link t-mono">Capabilities</a>
          <a href="/#stack" className="nav__link t-mono">Stack</a>
        </nav>

        <Link
          to="/agents"
          className={`nav__cta t-mono ${pathname === '/agents' ? 'is-active' : ''}`}
        >
          <span className="dot" />
          Agents
        </Link>
      </div>
    </header>
  );
}
