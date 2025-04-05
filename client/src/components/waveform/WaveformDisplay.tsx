import React, { useEffect, useRef } from "react";

type WaveformDisplayProps = {
	audioFile: File | null;
	color?: string;
};

const margin = 10;
const chunkSize = 50;

const WaveformDisplay: React.FC<WaveformDisplayProps> = ({ audioFile, color }) => {
	const canvasRef = useRef<HTMLCanvasElement | null>(null);

	useEffect(() => {
		if (!audioFile || !canvasRef.current) return;
		const drawToCanvas = async () => {
			const canvas = canvasRef.current;
			if (!canvas) return;
			const ctx = canvas.getContext("2d");
			if (!ctx) return;

			const ac = new AudioContext();
			const buffer = await audioFile.arrayBuffer();
			const audioBuffer = await ac.decodeAudioData(buffer);
			const float32Array = audioBuffer.getChannelData(0);

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

			// DRAW: Rounded background
			const radius = 10;
			ctx.clearRect(0, 0, canvas.width, canvas.height);
			ctx.fillStyle = "#8D98C1";
			ctx.beginPath();
			ctx.moveTo(radius, 0);
			ctx.lineTo(canvas.width - radius, 0);
			ctx.quadraticCurveTo(canvas.width, 0, canvas.width, radius);
			ctx.lineTo(canvas.width, canvas.height - radius);
			ctx.quadraticCurveTo(canvas.width, canvas.height, canvas.width - radius, canvas.height);
			ctx.lineTo(radius, canvas.height);
			ctx.quadraticCurveTo(0, canvas.height, 0, canvas.height - radius);
			ctx.lineTo(0, radius);
			ctx.quadraticCurveTo(0, 0, radius, 0);
			ctx.closePath();
			ctx.fill();

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
	}, [audioFile, color]);

	return (
		<canvas ref={canvasRef} height={200} />
	);
};

export default WaveformDisplay;
