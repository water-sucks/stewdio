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
			if (!canvas) return; // additional safeguard
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

			ctx.clearRect(0, 0, canvas.width, canvas.height);
			for (let index in array) {
				ctx.strokeStyle = color || "#000";
				ctx.beginPath();
				ctx.moveTo(margin + Number(index), centerHeight - array[index] * scaleFactor);
				ctx.lineTo(margin + Number(index), centerHeight + array[index] * scaleFactor);
				ctx.stroke();
			}
		};

		drawToCanvas();
	}, [audioFile, color]);

	return <canvas ref={canvasRef} height={200} />;
};

export default WaveformDisplay;
