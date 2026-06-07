import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { Sparkles } from '@react-three/drei';
import * as THREE from 'three';

/*
  Visual metaphor: a live threat (the pulsing red core) held inside a
  containment lattice (the rotating wireframe cages) while HydraNet watches
  it (the radar sweep + drifting captured-data particles).
*/

function Core() {
  const mesh = useRef();
  const matRef = useRef();

  useFrame((state) => {
    const t = state.clock.elapsedTime;
    const pulse = 0.5 + 0.5 * Math.sin(t * 2.1);
    if (matRef.current) {
      matRef.current.emissiveIntensity = 1.4 + pulse * 1.8;
    }
    if (mesh.current) {
      const s = 1 + pulse * 0.04;
      mesh.current.scale.setScalar(s);
      mesh.current.rotation.y = t * 0.25;
      mesh.current.rotation.x = t * 0.12;
    }
  });

  return (
    <mesh ref={mesh}>
      <icosahedronGeometry args={[1, 3]} />
      <meshStandardMaterial
        ref={matRef}
        color="#3a0a02"
        emissive="#ff3814"
        emissiveIntensity={2}
        roughness={0.35}
        metalness={0.6}
        flatShading
      />
    </mesh>
  );
}

function GlowShells() {
  const group = useRef();
  useFrame((state) => {
    const t = state.clock.elapsedTime;
    const pulse = 0.5 + 0.5 * Math.sin(t * 2.1);
    if (group.current) {
      group.current.children.forEach((m, i) => {
        const base = 1.25 + i * 0.4;
        m.scale.setScalar(base + pulse * 0.12);
        m.material.opacity = (0.12 - i * 0.035) * (0.6 + pulse * 0.6);
      });
    }
  });

  return (
    <group ref={group}>
      {[0, 1, 2].map((i) => (
        <mesh key={i}>
          <sphereGeometry args={[1, 32, 32]} />
          <meshBasicMaterial
            color="#ff3814"
            transparent
            opacity={0.1}
            blending={THREE.AdditiveBlending}
            depthWrite={false}
          />
        </mesh>
      ))}
    </group>
  );
}

function Cage() {
  const outer = useRef();
  const mid = useRef();

  useFrame((_, delta) => {
    if (outer.current) {
      outer.current.rotation.y += delta * 0.14;
      outer.current.rotation.x += delta * 0.045;
    }
    if (mid.current) {
      mid.current.rotation.y -= delta * 0.22;
      mid.current.rotation.z += delta * 0.05;
    }
  });

  return (
    <group>
      <mesh ref={outer} scale={2.65}>
        <icosahedronGeometry args={[1, 1]} />
        <meshBasicMaterial color="#c13b2a" wireframe transparent opacity={0.55} />
      </mesh>
      <mesh ref={mid} scale={2.1}>
        <dodecahedronGeometry args={[1, 0]} />
        <meshBasicMaterial color="#8a2616" wireframe transparent opacity={0.65} />
      </mesh>
    </group>
  );
}

function RadarSweep() {
  const ring = useRef();
  useFrame((_, delta) => {
    if (ring.current) {
      ring.current.rotation.z += delta * 0.9;
      ring.current.rotation.x = Math.PI / 2.3;
    }
  });
  return (
    <mesh ref={ring} scale={3.05}>
      <torusGeometry args={[1, 0.006, 8, 120]} />
      <meshBasicMaterial color="#ff8c42" transparent opacity={0.5} />
    </mesh>
  );
}

export default function CagedCore() {
  const drift = useRef();
  useFrame((state) => {
    if (drift.current) {
      // gentle breathing tilt so the whole rig never sits perfectly still
      drift.current.rotation.y = Math.sin(state.clock.elapsedTime * 0.15) * 0.25;
    }
  });

  return (
    <group ref={drift} scale={1.05}>
      <Core />
      <GlowShells />
      <Cage />
      <RadarSweep />
      <Sparkles count={70} scale={7} size={2.4} speed={0.28} color="#ff8c42" opacity={0.7} />
      <pointLight position={[0, 0, 0]} color="#ff3814" intensity={9} distance={9} decay={2} />
    </group>
  );
}
