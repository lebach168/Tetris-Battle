import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'

const AvatarSize = ({size}: {size: number}) => {
  if (size < 5 || size >7 || size == null) {
    size = 5
  }
    
  return (
    <Avatar className={`size-${size}`}>
      <AvatarImage src='https://cdn.shadcnstudio.com/ss-assets/avatar/avatar-5.png' alt='' />
      <AvatarFallback className='text-xs'>HR</AvatarFallback>
    </Avatar>
  )
}

export default AvatarSize
