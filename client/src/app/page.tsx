'use client';
import React, { useState } from 'react';
import FileUpload from '../components/FileSelector';
import WaveformCanvas from '../components/WaveformDisplay';

const Page = () => {
	const [audioFile1, setAudioFile1] = useState<File | null>(null);
	const [audioFile2, setAudioFile2] = useState<File | null>(null);

	return (
		<div style={{ display: 'flex', height: '100vh' }}>
			{/* Left column - file uploads */}
			<div style={{
				width: '250px',
				borderRight: '1px solid #ccc',
				overflowY: 'auto',
				padding: '1rem',
			}}>
				<h3>Select Files</h3>
				<FileUpload onFileSelect={setAudioFile1} />
				<FileUpload onFileSelect={setAudioFile2} />
				{/* You can add more FileUpload components as needed */}
			</div>

			{/* Right column - waveform display */}
			<div style={{
				flex: 1,
				overflowY: 'auto',
				padding: '1rem',
			}}>
				{audioFile1 && <WaveformCanvas audioFile={audioFile1} color={'white'} />}
				{audioFile2 && <WaveformCanvas audioFile={audioFile2} color={'red'} />}
			</div>
		</div>
	);
};

export default Page;
