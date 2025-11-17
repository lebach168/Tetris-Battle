import React from 'react';

type BlockProps = {
  type?: string; // 'ghost' \ none
  value: number;
  size?: number;
  opponent?: boolean;
};

const Cell: React.FC<BlockProps> = ({ type, value, size = 20, opponent }) => {
  const borderThickness = 3;
  const innerSize = size - borderThickness;

  // Base block với border đen cạnh dưới + phải
  const baseStyle = {
    width: `${size}px`,
    height: `${size}px`,
    boxSizing: 'border-box' as const,
    position: 'relative' as const,
    borderRight: `${borderThickness}px solid black`,
    borderBottom: `${borderThickness}px solid black`,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  };

  // Inner block
  const innerStyle = {
    width: `${innerSize}px`,
    height: `${innerSize}px`,
    position: 'relative' as const,
    boxSizing: 'border-box' as const,
  };

  if (type === 'ghost') {
    return (
      <div style={baseStyle}>
        <div
          style={{
            ...innerStyle,
            backgroundColor: 'transparent',
            border: '2px dashed #888',
          }}
        />
      </div>
    );
  }
  if (value === 0) {
    return <div style={{ ...baseStyle, borderRight: 'none', borderBottom: 'none' }} />;
  }
  const renderPixels = () => (
    <>
      <div
        style={{
          width: '2px',
          height: '2px',
          backgroundColor: 'white',
          position: 'absolute',
          top: '0px',
          left: '0px',
        }}
      />
      <div
        style={{
          width: '4px',
          height: '2px',
          backgroundColor: 'white',
          position: 'absolute',
          top: '2px',
          left: '2px',
        }}
      />
      <div
        style={{
          width: '2px',
          height: '4px',
          backgroundColor: 'white',
          position: 'absolute',
          top: '2px',
          left: '2px',
        }}
      />
    </>
  );

  let bg = '';
  let border = '';

  switch (value) {
    case 1:
    case 2:
    case 3:
      bg = 'white';
      border = !opponent ? '2px solid #58f898' : '2px solid #f83800';
      break;
    case 4:
    case 5:
      bg = !opponent ? '#6888fc' : '#fca044'; // blue : '#6888fc' '#3366FF'
      break;
    case 6:
    case 7:
      bg = !opponent ? '#58f898' : '#f83800'; //green : '#58f898' '#B0FC38'  '#A1FF59' '#00CC66'
      break;
    case 8:
      bg = '#999999'; // hoặc ##bcbcbc #7c7c7c #787878
      break;
    default:
      bg = 'transparent';
      border = 'none';
  }

  return (
    <div style={baseStyle}>
      <div
        style={{
          ...innerStyle,
          backgroundColor: bg,
          border: border || 'none',
        }}
      >
        <div>
          {value > 3 ? (
            renderPixels()
          ) : (
            <div
              style={{
                width: '2px',
                height: '2px',
                backgroundColor: 'white',
                position: 'absolute',
                top: '0px',
                left: '0px',
                zIndex: 1,
                transform: 'translate(-2px, -2px)',
              }}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default React.memo(Cell);
