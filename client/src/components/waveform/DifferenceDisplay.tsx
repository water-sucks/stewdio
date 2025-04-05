import React, { useEffect, useRef } from "react";

type WaveformDifferenceProps = {
	audioFile1: File | null;
	audioFile2: File | null;
	positiveColor?: string; // Color used when file1 is SHORTER
	negativeColor?: string; // Color used when file1 is LONGER
	equalLengthPositiveColor?: string; // Color for positive diff when lengths are equal
	equalLengthNegativeColor?: string; // Color for negative diff when lengths are equal
	backgroundColor?: string; // Optional background color
};

const margin = 10;
const chunkSize = 50;

const WaveformDifference: React.FC<WaveformDifferenceProps> = ({
	audioFile1,
	audioFile2,
	positiveColor = "#00FF00", // Default Green (File 1 shorter)
	negativeColor = "#FF0000", // Default Red (File 1 longer)
	equalLengthPositiveColor = "#00FF00", // Default Green for equal length positive diff
	equalLengthNegativeColor = "#FF0000", // Default Red for equal length negative diff
	backgroundColor = "#444444", // Default Dark Background
}) => {
	const canvasRef = useRef<HTMLCanvasElement | null>(null);

	useEffect(() => {
		// Guard clause: Exit if files or canvas are not ready
		if (!audioFile1 || !audioFile2 || !canvasRef.current) return;

		const drawDifference = async () => {
			const canvas = canvasRef.current;
			if (!canvas) return;
			const ctx = canvas.getContext("2d");
			if (!ctx) return;

			try {
				const ac = new AudioContext();

				// --- Audio Data Loading and Processing ---
				const buffer1 = await audioFile1.arrayBuffer();
				const audioBuffer1 = await ac.decodeAudioData(buffer1);
				const float32Array1 = audioBuffer1.getChannelData(0);

				const buffer2 = await audioFile2.arrayBuffer();
				const audioBuffer2 = await ac.decodeAudioData(buffer2);
				const float32Array2 = audioBuffer2.getChannelData(0);

				// --- Determine Lengths and Dominant Color ---
				const length1 = float32Array1.length;
				const length2 = float32Array2.length;
				let dominantColor: string | null = null;
				let useLengthBasedColoring = true;

				if (length1 < length2) {
					dominantColor = positiveColor; // File 1 is shorter, use positive color
				} else if (length1 > length2) {
					dominantColor = negativeColor; // File 1 is longer, use negative color
				} else {
					// Lengths are equal, use standard positive/negative coloring based on difference value
					useLengthBasedColoring = false;
				}

				// Determine the maximum length for comparison
				const maxLength = Math.max(length1, length2);

				// Calculate the sample-by-sample difference
				const differenceArray = new Float32Array(maxLength);
				for (let i = 0; i < maxLength; i++) {
					const val1 = float32Array1[i] || 0;
					const val2 = float32Array2[i] || 0;
					differenceArray[i] = val1 - val2;
				}

				// --- Chunking to find Positive and Negative Peaks ---
				// This logic remains the same, finding the actual peaks
				const positivePeaks: number[] = [];
				const negativePeaks: number[] = [];
				let i = 0;
				while (i < maxLength) {
					const chunkEnd = Math.min(i + chunkSize, maxLength);
					const chunk = differenceArray.slice(i, chunkEnd);
					i += chunkSize;

					if (chunk.length > 0) {
						let maxPositive = 0;
						let minNegative = 0;
						for (let j = 0; j < chunk.length; j++) {
							const value = chunk[j];
							if (value > 0) {
								maxPositive = Math.max(maxPositive, value);
							} else if (value < 0) {
								minNegative = Math.min(minNegative, value);
							}
						}
						positivePeaks.push(maxPositive);
						negativePeaks.push(minNegative);
					} else {
						positivePeaks.push(0);
						negativePeaks.push(0);
					}
				}

				// --- Canvas Drawing ---
				const numChunks = positivePeaks.length;
				const width = numChunks + margin * 2;
				const height = canvas.height;
				const centerHeight = Math.ceil(height / 2);
				const scaleFactor = (height - margin * 2) / 2;

				canvas.width = width;

				// Draw background
				if (backgroundColor) {
					const radius = 10;
					ctx.fillStyle = backgroundColor;
					ctx.beginPath();
					// ... (background drawing code remains the same) ...
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
				} else {
					ctx.clearRect(0, 0, canvas.width, canvas.height);
				}

				ctx.lineWidth = 1;

				// Draw the positive and negative peaks using the determined color logic
				for (let index = 0; index < numChunks; index++) {
					const x = margin + index;
					const positivePeak = positivePeaks[index];
					const negativePeak = negativePeaks[index]; // This value is <= 0

					// Draw positive peak portion
					if (positivePeak > 0) {
						const yPositive = positivePeak * scaleFactor;
						// Determine color: Use dominant if length-based, else use specific positive color
						ctx.strokeStyle = useLengthBasedColoring && dominantColor ? dominantColor : equalLengthPositiveColor;
						ctx.beginPath();
						ctx.moveTo(x, centerHeight);
						ctx.lineTo(x, centerHeight - yPositive); // Upwards
						ctx.stroke();
					}

					// Draw negative peak portion
					if (negativePeak < 0) {
						const yNegative = negativePeak * scaleFactor; // This will be negative
						// Determine color: Use dominant if length-based, else use specific negative color
						ctx.strokeStyle = useLengthBasedColoring && dominantColor ? dominantColor : equalLengthNegativeColor;
						ctx.beginPath();
						ctx.moveTo(x, centerHeight);
						ctx.lineTo(x, centerHeight - yNegative); // Downwards
						ctx.stroke();
					}
				}

			} catch (error) {
				console.error("Error processing or drawing waveform difference:", error);
				const canvas = canvasRef.current;
				if (canvas) {
					const ctx = canvas.getContext("2d");
					if (ctx) {
						ctx.clearRect(0, 0, canvas.width, canvas.height);
						ctx.fillStyle = "red";
						ctx.font = "12px sans-serif";
						ctx.textAlign = "center";
						ctx.fillText("Error loading/processing audio", canvas.width / 2, canvas.height / 2);
					}
				}
			} finally {
				// await ac.close(); // Consider AudioContext lifecycle
			}
		};

		drawDifference();

	}, [
		audioFile1,
		audioFile2,
		positiveColor,
		negativeColor,
		equalLengthPositiveColor, // Add new props to dependency array
		equalLengthNegativeColor, // Add new props to dependency array
		backgroundColor
	]);

	return <canvas ref={canvasRef} height={200} />;
};

export default WaveformDifference;
