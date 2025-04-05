'use client';
import React from 'react';
import WaveformCanvas from '../components/WaveformDisplay';
import FileSelector, { AudioFileItem } from '../components/FileSelector';

const Page = () => {
	const [selectedFiles, setSelectedFiles] = React.useState<AudioFileItem[]>([]);

	return (
		<div style={{ display: 'flex', height: '100vh' }}>
			{/* Left - Playlist */}
			<FileSelector onFileSelect={setSelectedFiles} />

			{/* Right - Waveform display */}
			<div style={{
				flex: 1,
				display: 'flex',
				flexDirection: 'column',
				padding: '1rem',
				overflowY: 'auto',
				gap: '1rem',
			}}>
				{selectedFiles.map((item, index) => (
					<div key={item.id} style={{ flex: 1, minHeight: '50%' }}>
						<WaveformCanvas
							audioFile={item.file}
							color={index === 0 ? '#2B5360' : '#764C7A'}
						/>
					</div>
				))}
			</div>
		</div>
	);
};

export default Page;
