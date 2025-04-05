'use client';
import React, { useState, useEffect } from 'react';

type FileSelectorProps = {
	onFileSelect: (files: AudioFileItem[]) => void;
};

export type AudioFileItem = {
	id: string;
	name: string;
	url: string; // Now contains URL instead of File
};

const FileSelector: React.FC<FileSelectorProps> = ({ onFileSelect = () => { } }) => {
	const [audioFiles, setAudioFiles] = useState<AudioFileItem[]>([]);
	const [selectedFiles, setSelectedFiles] = useState<AudioFileItem[]>([]);

	useEffect(() => {
		const fetchFiles = async () => {
			try {
				const response = await fetch('http://localhost:6969/api/v1/projects');
				const data = await response.json();

				// Assuming API returns something like: [{ id, name, url }, ...]
				const audioItems: AudioFileItem[] = data.map((item: any) => ({
					id: item.id,
					name: item.name,
					url: item.url,
				}));

				setAudioFiles(audioItems);
			} catch (err) {
				console.error('Failed to fetch audio files:', err);
			}
		};

		fetchFiles();
	}, []);

	const handleSelect = (file: AudioFileItem) => {
		setSelectedFiles((prev) => {
			let updated: AudioFileItem[];

			if (prev.find((f) => f.id === file.id)) {
				updated = prev.filter((f) => f.id !== file.id);
			} else if (prev.length < 2) {
				updated = [...prev, file];
			} else {
				updated = [prev[1], file];
			}

			onFileSelect(updated);
			return updated;
		});
	};

	return (
		<div style={{
			width: '300px',
			borderRight: '1px solid #ccc',
			display: 'flex',
			flexDirection: 'column',
			padding: '1rem',
		}}>
			<h3>Available Files</h3>

			<div style={{ overflowY: 'auto', flex: 1 }}>
				{audioFiles.map((item) => (
					<div
						key={item.id}
						onClick={() => handleSelect(item)}
						style={{
							padding: '0.5rem',
							marginBottom: '0.5rem',
							cursor: 'pointer',
							border: selectedFiles.find((f) => f.id === item.id)
								? '2px solid #0070f3'
								: '1px solid #ccc',
							borderRadius: '4px',
							backgroundColor: selectedFiles.find((f) => f.id === item.id)
								? '#e6f0ff'
								: '#f9f9f9',
						}}
					>
						{item.name}
					</div>
				))}
			</div>
		</div>
	);
};

export default FileSelector;
