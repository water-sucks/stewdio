import React, { useEffect, useRef } from "react";

type WaveformDisplayProps = {
	float32Array: Float32Array;
	color?: string;
};

const margin = 10;
const chunkSize = 50;

const Waveform: React.FC<WaveformDisplayProps> = ({ float32Array, color }) => {
	const canvasRef = useRef<HTMLCanvasElement | null>(null);

	useEffect(() => {
		if (!canvasRef.current) return;
		const drawToCanvas = async () => {
			const canvas = canvasRef.current;
			if (!canvas) return;
			const ctx = canvas.getContext("2d");
			if (!ctx) return;

			const array: number[] = [];
			let i = 0;
			const length = float32Array.length;
			while (i < length) {
				array.push(
					float32Array.slice(i, i += chunkSize).reduce((total, value) => {
						return Math.max(total, Math.abs(value));
					})
				);
			}

			const width = Math.ceil(float32Array.length / chunkSize + margin * 2);
			const height = canvas.height;
			const centerHeight = Math.ceil(height / 2);
			const scaleFactor = (height - margin * 2) / 2;

			canvas.width = width;

			// DRAW: Waveform
			ctx.strokeStyle = color || "#000";
			for (let index in array) {
				const x = margin + Number(index);
				const y = array[index] * scaleFactor;
				ctx.beginPath();
				ctx.moveTo(x, centerHeight - y);
				ctx.lineTo(x, centerHeight + y);
				ctx.stroke();
			}
		};

		drawToCanvas();
	}, [float32Array, color]);

	return (
		<canvas
			ref={canvasRef}
			style={{
				height: "200",
				borderRadius: "10px",
				backgroundColor: "#8D98C1",
			}}
		/>
	);
};

export default Waveform;
