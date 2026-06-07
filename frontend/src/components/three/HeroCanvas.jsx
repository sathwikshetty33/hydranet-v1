import { Suspense } from 'react';
import { Canvas } from '@react-three/fiber';
import CagedCore from './CagedCore';

export default function HeroCanvas() {
  return (
    <Canvas
      camera={{ position: [0, 0.4, 7.2], fov: 42 }}
      dpr={[1, 2]}
      gl={{ antialias: true, alpha: true, powerPreference: 'high-performance' }}
      style={{ width: '100%', height: '100%' }}
    >
      <fog attach="fog" args={['#0a0908', 7, 17]} />
      <ambientLight intensity={0.18} />
      <directionalLight position={[4, 6, 5]} intensity={0.4} color="#ffb088" />
      <Suspense fallback={null}>
        <CagedCore />
      </Suspense>
    </Canvas>
  );
}
