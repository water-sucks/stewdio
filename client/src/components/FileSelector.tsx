'use client';
import React, { useState } from 'react';

type FileSelectorProps = {
	onFileSelect: (files: AudioFileItem[]) => void;
};

export type AudioFileItem = {
	id: string;
	name: string;
	file: File;
};

const FileSelector: React.FC<FileSelectorProps> = ({ onFileSelect = () => { } }) => {
	const [audioFiles, setAudioFiles] = useState<AudioFileItem[]>([]);
	const [selectedFiles, setSelectedFiles] = useState<AudioFileItem[]>([]);

	const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
		const files = Array.from(e.target.files || []);
		const newItems = files.map((file) => ({
			id: crypto.randomUUID(),
			name: file.name,
			file,
		}));
		setAudioFiles((prev) => [...prev, ...newItems]);
	};

	const handleSelect = (file: AudioFileItem) => {
		setSelectedFiles((prev) => {
			let updated: AudioFileItem[];

			if (prev.find((f) => f.id === file.id)) {
				// Deselect
				updated = prev.filter((f) => f.id !== file.id);
			} else if (prev.length < 2) {
				// Select new one
				updated = [...prev, file];
			} else {
				// Replace the first selected file
				updated = [prev[1], file];
			}

			onFileSelect(updated); // Pass to parent
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
			<input
				type="file"
				accept="audio/*"
				multiple
				onChange={handleFileUpload}
				style={{ marginBottom: '1rem' }}
			/>

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
