import { useCallback, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';

import { ChevronLeftIcon, ChevronRightIcon, PlusIcon, RefreshCcw } from 'lucide-react';

import type { ColumnDef, SortingState, VisibilityState } from '@tanstack/react-table';
import {
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
} from '@/components/ui/pagination';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

import { cn } from '@/utils/utils';
import { usePagination } from '@/hooks/use-pagination';
import AvatarSize from '../Unit/SizeAvatar';

type Player = { id: string; name: string; avatarUrl?: string };
type Room = { id: string; name: string; players: Player[] };

const data: Room[] = [
  { id: '1', name: 'ABC12', players: [{ id: 'u1', name: 'Alice' }] },
  {
    id: '2',
    name: 'DEF34',
    players: [
      { id: 'u2', name: 'Bob' },
      { id: 'u3', name: 'Cindy' },
    ],
  },
  { id: '3', name: 'XXXXX', players: [] },
];

const avatarCell = (player?: Player) => {
  if (!player) return <span className="text-muted-foreground">…waiting</span>;
  return (
    <div className="flex items-center gap-1">
      {/* <div className='h-8 w-8 rounded-full bg-muted text-muted-foreground grid place-items-center'>{initial}</div>
      <span className='truncate max-w-[10rem]' title={player.name}>{player.name}</span> */}
      <AvatarSize size={5} />
      <span className="truncate max-w-[10rem]" title={player.name}>
        {player.name}
      </span>
    </div>
  );
};

// Simulate a POST request to create a new room
const createNewRoomApi = async () => {
  // Simulate network delay
  await new Promise((res) => setTimeout(res, 500));
  // Fixed response
  return {
    room: {
      id: 'ABC00',
      players: [],
    },
    ws_url: 'ws://localhost:8080/ws/match?roomid=ABC00&playerid=anon123',
  };
};

// Simulate a POST request to join a room
const joinRoomApi = async (roomId: string) => {
  // Simulate network delay
  await new Promise((res) => setTimeout(res, 500));
  // Fixed response
  return {
    room: {
      id: roomId,
      players: [{ id: 'anon123' }],
    },
    ws_url: `ws://localhost:8080/ws/match?roomid=${roomId}&playerid=anon456`,
  };
};

// columns will be created inside the component to access handlers safely

const ListRoom = () => {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});
  const [nameFilter, setNameFilter] = useState('');

  const filteredData = useMemo(() => {
    const q = nameFilter.trim().toLowerCase();
    if (!q) return data;
    return data.filter((r) => r.name.toLowerCase().includes(q));
  }, [nameFilter]);

  const navigate = useNavigate();

  // Handlers need to be defined before columns so they are in scope
  const handleCreateNewRoom = useCallback(async () => {
    const res = await createNewRoomApi();
    const playerId = new URL(res.ws_url).searchParams.get('playerid');
    if (playerId) {
      navigate({ to: `/match/${playerId}` });
    }
  }, [navigate]);

  const handleJoinRoom = useCallback(
    async (roomId: string) => {
      const res = await joinRoomApi(roomId);
      const playerId = new URL(res.ws_url).searchParams.get('playerid');
      if (playerId) {
        navigate({ to: `/match/${playerId}` });
      }
    },
    [navigate],
  );

  const columns: ColumnDef<Room>[] = useMemo(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <div className="font-medium text-left">{row.getValue('name') as string}</div>
        ),
      },
      {
        id: 'status',
        header: 'Status',
        cell: ({ row }) => {
          const players = row.original.players?.length ?? 0;
          const total = 2;
          const isFull = players >= total;
          const text = isFull ? 'Full' : `${players}/${total}`;
          const styles = isFull
            ? 'bg-destructive/10 text-destructive'
            : 'bg-green-600/10 text-green-600 dark:bg-green-400/10 dark:text-green-400';
          return <Badge className={cn('rounded-full border-none', styles)}>{text}</Badge>;
        },
      },
      {
        id: 'player1',
        header: 'Player 1',
        cell: ({ row }) => avatarCell(row.original.players?.[0]),
      },
      {
        id: 'player2',
        header: 'Player 2',
        cell: ({ row }) => avatarCell(row.original.players?.[1]),
      },
      {
        id: 'join',
        cell: ({ row }) => {
          const players = row.original.players?.length ?? 0;
          const total = 2;
          const isFull = players >= total;
          return (
            <div className="text-right font-medium">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleJoinRoom(row.original.id)}
                className="gap-2"
                disabled={isFull}
              >
                Join
              </Button>
            </div>
          );
        },
      },
    ],
    [handleJoinRoom],
  );

  const table = useReactTable({
    data: filteredData,
    columns,
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
    },
  });

  const currentPage = table.getState().pagination?.pageIndex
    ? table.getState().pagination.pageIndex + 1
    : 1;
  const totalPages = table.getPageCount();
  const { pages, showLeftEllipsis, showRightEllipsis } = usePagination({
    currentPage,
    totalPages,
    paginationItemsToDisplay: 5,
  });

  return (
    <div className="w-full">
      <div className="flex justify-between gap-2 pb-4 max-sm:flex-col sm:items-center">
        <div className="flex items-center space-x-2">
          <Input
            placeholder="Search room name..."
            value={nameFilter}
            onChange={(event) => setNameFilter(String(event.target.value))}
            className="max-w-sm"
          />
          <Button
            variant="outline"
            size="icon"
            onClick={() => window.location.reload()}
            className="gap-2"
          >
            <RefreshCcw className="h-4 w-4" />
          </Button>
        </div>

        <div className="flex items-center space-x-2">
          <Button size="sm" onClick={handleCreateNewRoom} className="gap-2">
            <PlusIcon className="h-4 w-4" /> Create
          </Button>
        </div>
      </div>
      <div className="rounded-md border">
        <Table className="table-fixed">
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  const width = header.getSize?.() ?? undefined;
                  return (
                    <TableHead key={header.id} style={width ? { width } : undefined}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      key={cell.id}
                      style={
                        cell.column.getSize?.() ? { width: cell.column.getSize?.() } : undefined
                      }
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="grow">
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <Button
                size="icon"
                variant="outline"
                className="disabled:pointer-events-none disabled:opacity-50"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
                aria-label="Go to previous page"
              >
                <ChevronLeftIcon size={16} aria-hidden="true" />
              </Button>
            </PaginationItem>

            {showLeftEllipsis && (
              <PaginationItem>
                <PaginationEllipsis />
              </PaginationItem>
            )}

            {pages.map((page) => {
              const isActive = page === table.getState().pagination.pageIndex + 1;

              return (
                <PaginationItem key={page}>
                  <Button
                    size="icon"
                    variant={`${isActive ? 'outline' : 'ghost'}`}
                    onClick={() => table.setPageIndex(page - 1)}
                    aria-current={isActive ? 'page' : undefined}
                  >
                    {page}
                  </Button>
                </PaginationItem>
              );
            })}

            {showRightEllipsis && (
              <PaginationItem>
                <PaginationEllipsis />
              </PaginationItem>
            )}

            <PaginationItem>
              <Button
                size="icon"
                variant="outline"
                className="disabled:pointer-events-none disabled:opacity-50"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
                aria-label="Go to next page"
              >
                <ChevronRightIcon size={16} aria-hidden="true" />
              </Button>
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </div>
    </div>
  );
};

export default ListRoom;
